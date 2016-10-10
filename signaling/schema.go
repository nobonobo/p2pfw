package signaling

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// N ...
	N = 1024
	// TIMEOUT ...
	TIMEOUT = 30 * time.Second
)

// Member ...
type Member struct {
	sync.RWMutex
	UserID string
	event  chan *Event
	timer  *time.Timer
}

// Push ...
func (m *Member) Push(event *Event) {
	m.RLock()
	ch := m.event
	m.RUnlock()
	for {
		select {
		case ch <- event:
			log.Println("push:", m.UserID, event)
			return
		default:
		}
		_, ok := <-ch
		if !ok {
			return
		}
	}
}

// Pop ...
func (m *Member) Pop() <-chan *Event {
	m.RLock()
	ch := m.event
	m.RUnlock()
	return ch
}

// Reset ...
func (m *Member) Reset() {
	m.Lock()
	m.timer.Reset(TIMEOUT)
	m.Unlock()
}

// Close ...
func (m *Member) Close() {
	m.Lock()
	m.timer.Stop()
	close(m.event)
	m.Unlock()
}

// Room ...
type Room struct {
	name      string
	owner     string
	preshared string
	sync.RWMutex
	members map[string]*Member
	check   func()
	locked  bool
}

// NewRoom ...
func NewRoom(name, owner, preshared string) *Room {
	room := &Room{
		name:      name,
		owner:     owner,
		preshared: preshared,
		members:   map[string]*Member{},
	}
	room.Join(owner)
	return room
}

// Name ...
func (r *Room) Name() string {
	return r.name
}

// Preshared ...
func (r *Room) Preshared() string {
	return r.preshared
}

// Owner ...
func (r *Room) Owner() string {
	return r.owner
}

// SetCheckFunc ...
func (r *Room) SetCheckFunc(check func()) {
	r.check = check
}

// Locked ...
func (r *Room) Locked() bool {
	r.RLock()
	defer r.RUnlock()
	return r.locked
}

// SetLocked ...
func (r *Room) SetLocked(b bool) {
	r.Lock()
	defer r.Unlock()
	r.locked = b
}

// Join ...
func (r *Room) Join(user string) error {
	r.Lock()
	defer r.Unlock()
	if m, ok := r.members[user]; ok {
		m.Reset()
		return nil
	}
	if r.locked {
		return fmt.Errorf("room is locked")
	}
	m := &Member{
		UserID: user,
		event:  make(chan *Event, N),
	}
	m.timer = time.AfterFunc(TIMEOUT, func() { r.Leave(user) })
	r.members[user] = m
	return nil
}

// Leave ...
func (r *Room) Leave(user string) error {
	defer func() {
		r.RLock()
		_, ok := r.members[r.owner]
		r.RUnlock()
		if !ok {
			r.check()
		}
	}()
	r.Lock()
	defer r.Unlock()
	m, ok := r.members[user]
	if !ok {
		return fmt.Errorf("not found user: %s", user)
	}
	delete(r.members, user)
	m.Close()
	return nil
}

// Get ...
func (r *Room) Get(user string) *Member {
	r.RLock()
	m := r.members[user]
	r.RUnlock()
	return m
}

// Iter ...
func (r *Room) Iter(fn func(m *Member)) {
	r.RLock()
	defer r.RUnlock()
	for _, m := range r.members {
		fn(m)
	}
}

// Send ...
func (r *Room) Send(msg Message) error {
	r.RLock()
	defer r.RUnlock()
	self := r.members[msg.UserID]
	if self == nil {
		return fmt.Errorf("you not a member: %s", msg.UserID)
	}
	self.Reset()
	if m := r.members[msg.Event.To]; msg.Event.To != "" && m != nil {
		m.Push(msg.Event)
	} else {
		for _, m := range r.members {
			if m != self {
				m.Push(msg.Event)
			}
		}
	}
	return nil
}

// Close ...
func (r *Room) Close() {
	r.Lock()
	defer r.Unlock()
	members := r.members
	r.members = map[string]*Member{}
	for _, m := range members {
		m.Close()
	}
}

// Request ...
type Request struct {
	RoomID    string
	UserID    string
	Preshared string
}

// Valid ...
func (r Request) Valid() error {
	if len(r.RoomID) == 0 {
		return fmt.Errorf("must set RoomID")
	}
	if len(r.UserID) == 0 {
		return fmt.Errorf("must set UserID")
	}
	return nil
}

// Message ...
type Message struct {
	Request
	Event *Event
}

// Members ...
type Members struct {
	Owner  string
	Member []string
}
