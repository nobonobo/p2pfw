package peerconn

import (
	"sync"

	"github.com/nobonobo/webrtc"
)

// Conn ...
type Conn struct {
	*webrtc.PeerConnection
	peer       string
	mu         sync.RWMutex
	candidates []*webrtc.IceCandidate
	datachans  map[string]*webrtc.DataChannel
}

// NewConn ...
func NewConn(peer string, pc *webrtc.PeerConnection) *Conn {
	c := &Conn{
		PeerConnection: pc,
		peer:           peer,
		candidates:     []*webrtc.IceCandidate{},
		datachans:      map[string]*webrtc.DataChannel{},
	}
	pc.OnDataChannel(c.SetDataChannel)
	return c
}

// SetDataChannel ...
func (p *Conn) SetDataChannel(dc *webrtc.DataChannel) {
	p.mu.Lock()
	p.datachans[dc.Label()] = dc
	p.mu.Unlock()
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
