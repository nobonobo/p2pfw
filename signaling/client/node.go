package client

import (
	"github.com/nobonobo/p2pfw/signaling"
)

// Dispatcher ...
type Dispatcher interface {
	Dispatch([]*signaling.Event)
}

// DispatcherFunc ...
type DispatcherFunc func([]*signaling.Event)

// Dispatch ...
func (fn DispatcherFunc) Dispatch(events []*signaling.Event) {
	fn(events)
}

// Node ...
type Node struct {
	r         signaling.Request
	rpcClient *Client
	closing   chan struct{}
	done      chan error
	err       error
}

// NewNode ...
func NewNode(config *Config) (*Node, error) {
	c, err := New(config)
	if err != nil {
		return nil, err
	}
	n := &Node{
		r:         config.Request,
		rpcClient: c,
	}
	n.done = make(chan error)
	close(n.done)
	return n, nil
}

func (n *Node) run(dispatchers ...Dispatcher) {
	defer close(n.done)
	for {
		if err := n.rpcClient.Call("Signaling.Join", n.r, None); err != nil {
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
			for _, d := range dispatchers {
				d.Dispatch(*events)
			}
		}
	}
}

// Room ...
func (n *Node) Room() string { return n.r.RoomID }

// User ...
func (n *Node) User() string { return n.r.UserID }

// Start ...
func (n *Node) Start(owner bool, dispatchers ...Dispatcher) error {
	if err := n.Stop(); err != nil {
		return err
	}
	if owner {
		if err := n.rpcClient.Call("Signaling.CreateRoom", n.r, None); err != nil {
			return err
		}
	} else {
		if err := n.rpcClient.Call("Signaling.Join", n.r, None); err != nil {
			return err
		}
	}
	n.closing = make(chan struct{})
	n.done = make(chan error, 1)
	go n.run(dispatchers...)
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

// Send ...
func (n *Node) Send(ev *signaling.Event) error {
	return n.rpcClient.Call("Signaling.Send",
		signaling.Message{Request: n.r, Event: ev},
		None,
	)
}
