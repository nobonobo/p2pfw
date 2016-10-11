package peerconn

import (
	"log"

	"github.com/nobonobo/p2pfw/signaling"
	"github.com/nobonobo/p2pfw/signaling/client"
	"github.com/nobonobo/webrtc"
)

// Node ...
type Node struct {
	r         signaling.Request
	rpcClient *client.Client
	closing   chan struct{}
	done      chan error
	err       error

	config  *webrtc.Configuration
	Clients *Connections // 接続元
	Servers *Connections // 接続先

	OnJoin           func(member string)
	OnLeave          func(member string)
	OnPeerConnection func(string, *Conn) error
}

// NewNode ...
func NewNode(dial *client.Config, config *webrtc.Configuration) (*Node, error) {
	c, err := client.New(dial)
	if err != nil {
		return nil, err
	}
	n := &Node{
		r:                dial.Request,
		rpcClient:        c,
		config:           config,
		Clients:          NewConnections(),
		Servers:          NewConnections(),
		OnJoin:           func(string) {},
		OnLeave:          func(string) {},
		OnPeerConnection: func(string, *Conn) error { return nil },
	}
	n.done = make(chan error)
	close(n.done)
	return n, nil
}

func (n *Node) run() {
	defer close(n.done)
	for {
		if err := n.rpcClient.Call("Signaling.Join", n.r, client.None); err != nil {
			n.done <- err
			return
		}
		events := new([]*signaling.Event)
		call := n.rpcClient.Go("Signaling.Pull", n.r, &events, nil)
		select {
		case <-n.closing:
			return
		case <-call.Done:
			if call.Error != nil {
				n.done <- call.Error
				return
			}
			n.dispatch(*events)
		}
	}
}

func (n *Node) dispatch(events []*signaling.Event) {
	for _, ev := range events {
		msg := ev.Get()
		log.Printf("recv from %s: %#v", ev.From, msg)
		switch v := msg.(type) {
		case *signaling.Join:
			n.OnJoin(v.Member)
		case *signaling.Leave:
			n.OnLeave(v.Member)
		case *Connect:
			pc, err := webrtc.NewPeerConnection(n.config)
			if err != nil {
				log.Printf("%s: %s", ev.From, err)
				break
			}
			conn := NewConn(ev.From, pc)
			conn.OnIceCandidate(func(ic *webrtc.IceCandidate) {
				if err := n.Send(ev.From, (*OfferCandidate)(ic)); err != nil {
					log.Printf("%s: %s", ev.From, err)
				}
			})
			conn.OnIceCandidateError(func() {
				log.Printf("%s: ice candidate failed", ev.From)
				if err := n.Send(ev.From, &OfferFailed{}); err != nil {
					log.Printf("%s: %s", ev.From, err)
				}
			})
			conn.OnIceGatheringStateChange(func(s string) {
				log.Printf("%s: ice gathering state change %q", ev.From, s)
				if s == "Complete" {
					if err := n.Send(ev.From, &OfferCompleted{}); err != nil {
						log.Printf("%s: %s", ev.From, err)
					}
				}
			})
			if err := n.OnPeerConnection(ev.From, conn); err != nil {
				log.Printf("%s: %s", ev.From, err)
				break
			}
			sdp, err := conn.CreateOffer()
			if err != nil {
				log.Printf("%s: %s", ev.From, err)
				break
			}
			if err := conn.SetLocalDescription(sdp); err != nil {
				log.Printf("%s: %s", ev.From, err)
				break
			}
			if err := n.Send(ev.From, (*Offer)(sdp)); err != nil {
				log.Printf("%s: %s", ev.From, err)
				break
			}
			n.Clients.Set(ev.From, conn)
		case *Offer:
			if conn := n.Servers.Get(ev.From); conn != nil {
				sdp := (*webrtc.SessionDescription)(v)
				if err := conn.SetRemoteDescription(sdp); err != nil {
					log.Printf("%s: %s", ev.From, err)
					break
				}
				sdp, err := conn.CreateAnswer()
				if err != nil {
					log.Printf("%s: %s", ev.From, err)
					break
				}
				if err := conn.SetLocalDescription(sdp); err != nil {
					log.Printf("%s: %s", ev.From, err)
					break
				}
				if err := n.Send(ev.From, (*Answer)(sdp)); err != nil {
					log.Printf("%s: %s", ev.From, err)
					break
				}
			}
		case *OfferCandidate:
			if conn := n.Servers.Get(ev.From); conn != nil {
				ic := (*webrtc.IceCandidate)(v)
				conn.AppendIceCandidate(ic)
			}
		case *OfferCompleted:
			if conn := n.Servers.Get(ev.From); conn != nil {
				if err := conn.ApplyIceCandidates(); err != nil {
					log.Printf("%s: %s", ev.From, err)
				}
			}
		case *OfferFailed:
			n.Servers.Del(ev.From)
		case *Answer:
			if conn := n.Clients.Get(ev.From); conn != nil {
				sdp := (*webrtc.SessionDescription)(v)
				if err := conn.SetRemoteDescription(sdp); err != nil {
					log.Printf("%s: %s", ev.From, err)
					break
				}
			}
		case *AnswerCandidate:
			if conn := n.Clients.Get(ev.From); conn != nil {
				ic := (*webrtc.IceCandidate)(v)
				conn.AppendIceCandidate(ic)
			}
		case *AnswerCompleted:
			if conn := n.Clients.Get(ev.From); conn != nil {
				if err := conn.ApplyIceCandidates(); err != nil {
					log.Printf("%s: %s", ev.From, err)
				}
			}
		case *AnswerFailed:
			n.Clients.Del(ev.From)
		default:
			log.Printf("%s: unsupported event %#v", ev.From, msg)
		}
	}
}

// Room ...
func (n *Node) Room() string { return n.r.RoomID }

// User ...
func (n *Node) User() string { return n.r.UserID }

// Start ...
func (n *Node) Start(owner bool) error {
	if err := n.Stop(); err != nil {
		return err
	}
	if owner {
		if err := n.rpcClient.Call("Signaling.CreateRoom", n.r, client.None); err != nil {
			return err
		}
	} else {
		if err := n.rpcClient.Call("Signaling.Join", n.r, client.None); err != nil {
			return err
		}
	}
	n.closing = make(chan struct{})
	n.done = make(chan error, 1)
	go n.run()
	return nil
}

// Stop ...
func (n *Node) Stop() error {
	select {
	case <-n.done:
		return nil
	default:
	}
	close(n.closing)
	n.err = <-n.done
	return n.err
}

// Close ...
func (n *Node) Close() error {
	existErr := n.Stop()
	n.Servers.Iter(func(_ string, c *Conn) {
		err := c.Close()
		if existErr == nil && err != nil {
			existErr = err
		}
	})
	n.Clients.Iter(func(_ string, c *Conn) {
		err := c.Close()
		if existErr == nil && err != nil {
			existErr = err
		}
	})
	return existErr
}

// Send ...
func (n *Node) Send(dest string, v signaling.Kinder) error {
	log.Printf("send to %s: %#v", dest, v)
	return n.rpcClient.Call("Signaling.Send",
		signaling.Message{Request: n.r, Event: signaling.New(n.r.UserID, dest, v)},
		client.None,
	)
}

// Members ...
func (n *Node) Members() (*signaling.Members, error) {
	var m *signaling.Members
	if err := n.rpcClient.Call("Signaling.Members", n.r, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Connect ...
func (n *Node) Connect(peer string) (*Conn, error) {
	pc, err := webrtc.NewPeerConnection(n.config)
	if err != nil {
		return nil, err
	}
	conn := NewConn(peer, pc)
	conn.OnIceCandidate(func(ic *webrtc.IceCandidate) {
		if err := n.Send(conn.Peer(), (*AnswerCandidate)(ic)); err != nil {
			log.Printf("%s: %s", conn.Peer(), err)
		}
	})
	conn.OnIceCandidateError(func() {
		log.Println("ice candidate failed:")
		if err := n.Send(conn.Peer(), &AnswerFailed{}); err != nil {
			log.Printf("%s: %s", conn.Peer(), err)
		}
	})
	conn.OnIceGatheringStateChange(func(s string) {
		log.Println("ice gathering state change:", s)
		if s == "Complete" {
			if err := n.Send(conn.Peer(), &AnswerCompleted{}); err != nil {
				log.Printf("%s: %s", conn.Peer(), err)
			}
		}
	})
	if err := n.Send(peer, &Connect{}); err != nil {
		return nil, err
	}
	n.Servers.Set(peer, conn)
	return conn, nil
}
