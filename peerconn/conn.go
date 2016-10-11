package peerconn

import (
	"sync"

	"github.com/nobonobo/webrtc"
)

// Conn ...
type Conn struct {
	*webrtc.PeerConnection
	peer          string
	mu            sync.RWMutex
	candidates    []*webrtc.IceCandidate
	datachans     map[string]*webrtc.DataChannel
	ondatachannel func(dc *webrtc.DataChannel)
}

// NewConn ...
func NewConn(peer string, pc *webrtc.PeerConnection) *Conn {
	p := &Conn{
		PeerConnection: pc,
		peer:           peer,
		candidates:     []*webrtc.IceCandidate{},
		datachans:      map[string]*webrtc.DataChannel{},
	}
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		p.SetDataChannel(dc)
		if p.ondatachannel != nil {
			p.ondatachannel(dc)
		}
	})
	return p
}

// OnDataChannel ...
func (p *Conn) OnDataChannel(fn func(dc *webrtc.DataChannel)) {
	p.ondatachannel = fn
}

// SetDataChannel ...
func (p *Conn) SetDataChannel(dc *webrtc.DataChannel) {
	p.mu.Lock()
	p.datachans[dc.Label()] = dc
	p.mu.Unlock()
}

// Close ...
func (p *Conn) Close() error {
	var err error
	p.mu.Lock()
	for _, dc := range p.datachans {
		if e := dc.Close(); e != nil && err == nil {
			err = e
		}
	}
	p.mu.Unlock()
	if e := p.PeerConnection.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

// Peer ...
func (p *Conn) Peer() string {
	return p.peer
}

// AppendIceCandidate ...
func (p *Conn) AppendIceCandidate(ic *webrtc.IceCandidate) {
	p.mu.Lock()
	p.candidates = append(p.candidates, ic)
	p.mu.Unlock()
}

// ApplyIceCandidates ...
func (p *Conn) ApplyIceCandidates() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, ic := range p.candidates {
		if err := p.PeerConnection.AddIceCandidate(ic); err != nil {
			return err
		}
	}
	return nil
}

// Connections ...
type Connections struct {
	sync.RWMutex
	m map[string]*Conn
}

// NewConnections ...
func NewConnections() *Connections {
	return &Connections{
		m: map[string]*Conn{},
	}
}

// Set ...
func (p *Connections) Set(user string, peer *Conn) {
	p.Lock()
	if old, ok := p.m[user]; ok {
		old.Close()
	}
	p.m[user] = peer
	p.Unlock()
}

// Del ...
func (p *Connections) Del(user string) {
	p.Lock()
	if old, ok := p.m[user]; ok {
		old.Close()
		delete(p.m, user)
	}
	p.Unlock()
}

// Get ...
func (p *Connections) Get(user string) *Conn {
	p.RLock()
	peer := p.m[user]
	p.RUnlock()
	return peer
}

// Iter ...
func (p *Connections) Iter(fn func(string, *Conn)) {
	p.RLock()
	defer p.RUnlock()
	for name, peer := range p.m {
		fn(name, peer)
	}
}
