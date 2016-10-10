package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"sync"
	"time"

	"github.com/rs/cors"
	"golang.org/x/net/websocket"

	jrpc "github.com/nobonobo/p2pfw/jsonrpc"
	"github.com/nobonobo/p2pfw/signaling"
)

// Signaling ...
type Signaling struct {
	mutex sync.RWMutex
	rooms map[string]*signaling.Room
}

// Pull ...
func (s *Signaling) Pull(req signaling.Request, events *[]*signaling.Event) (er error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	m := room.Get(req.UserID)
	if m == nil {
		return fmt.Errorf("not found member: %s", req.UserID)
	}
	*events = []*signaling.Event{}
	m.Reset()
	tm := time.NewTimer(3 * time.Second)
	select {
	case event, ok := <-m.Pop():
		if !ok {
			return nil
		}
		*events = append(*events, event)
		tm.Reset(3 * time.Second)
		for {
			select {
			case event, ok := <-m.Pop():
				if !ok {
					return nil
				}
				*events = append(*events, event)
				continue
			default:
			}
			return nil
		}
	case <-tm.C:
	}
	return nil
}

// CreateRoom ...
func (s *Signaling) CreateRoom(req signaling.Request, none *struct{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := req.Valid(); err != nil {
		return err
	}
	if room, ok := s.rooms[req.RoomID]; ok {
		if room.Owner() == req.UserID && room.Preshared() == req.Preshared {
			room.Join(req.UserID)
			return nil
		}
		return fmt.Errorf("room name duplicated: %s", req.RoomID)
	}
	room := signaling.NewRoom(
		req.RoomID,
		req.UserID,
		req.Preshared,
	)
	room.SetCheckFunc(func() {
		s.mutex.Lock()
		s.destroyRoom(room)
		s.mutex.Unlock()
	})
	s.rooms[room.Name()] = room
	log.Println("create room:", room.Name())
	return nil
}

func (s *Signaling) destroyRoom(room *signaling.Room) {
	delete(s.rooms, room.Name())
	room.Close()
	log.Println("destroy room:", room.Name())
}

// DestroyRoom ...
func (s *Signaling) DestroyRoom(req signaling.Request, none *struct{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	if room.Owner() != req.UserID {
		return fmt.Errorf("no permission: %s", req.UserID)
	}
	s.destroyRoom(room)
	return nil
}

// Join ...
func (s *Signaling) Join(req signaling.Request, none *struct{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	if err := room.Join(req.UserID); err != nil {
		return err
	}
	room.Send(signaling.New(req.UserID, "",
		&signaling.Join{Member: "req.UserID"},
	))
	return nil
}

// Leave ...
func (s *Signaling) Leave(req signaling.Request, none *struct{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	if err := room.Leave(req.UserID); err != nil {
		return err
	}
	room.Send(signaling.New(req.UserID, "",
		&signaling.Leave{Member: "req.UserID"},
	))
	return nil
}

// Locked ...
func (s *Signaling) Locked(req signaling.Request, locked *bool) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	if room.Get(req.UserID) == nil {
		return fmt.Errorf("you not a member: %s", req.UserID)
	}
	*locked = room.Locked()
	return nil
}

// SetLocked ...
type SetLocked struct {
	signaling.Request
	Locked bool
}

// SetLocked ...
func (s *Signaling) SetLocked(req *SetLocked, none *struct{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	if room.Get(req.UserID) == nil {
		return fmt.Errorf("you not a member: %s", req.UserID)
	}
	room.SetLocked(req.Locked)
	return nil
}

// Members ...
func (s *Signaling) Members(req signaling.Request, members *signaling.Members) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := req.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[req.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", req.RoomID)
	}
	if room.Preshared() != req.Preshared {
		return fmt.Errorf("mismatch preshared: %s", req.Preshared)
	}
	if room.Get(req.UserID) == nil {
		return fmt.Errorf("you not a member: %s", req.UserID)
	}
	room.Iter(func(m *signaling.Member) {
		if room.Owner() == m.UserID {
			(*members).Owner = m.UserID
			return
		}
		(*members).Member = append((*members).Member, m.UserID)
	})
	return nil
}

// Send ...
func (s *Signaling) Send(msg signaling.Message, none *struct{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if err := msg.Valid(); err != nil {
		return err
	}
	room, ok := s.rooms[msg.RoomID]
	if !ok {
		return fmt.Errorf("not found room: %s", msg.RoomID)
	}
	if room.Preshared() != msg.Preshared {
		return fmt.Errorf("mismatch preshared: %s", msg.Preshared)
	}
	return room.Send(msg)
}

func wsHandle(ws *websocket.Conn) {
	log.Println("connect:", ws.Request().RemoteAddr)
	defer log.Println("disconnect:", ws.Request().RemoteAddr)
	jsonrpc.ServeConn(ws)
}

func getStun(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, os.Getenv("STUN"))
}

func main() {
	rpc.Register(&Signaling{rooms: map[string]*signaling.Room{}})
	l, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/ws", websocket.Handler(wsHandle))
	http.Handle("/stun", cors.Default().Handler(
		http.HandlerFunc(getStun)),
	)
	http.Handle("/", cors.Default().Handler(jrpc.Handle))
	log.Println("signaling server:", l.Addr())
	if err := http.Serve(l, nil); err != nil {
		log.Fatalln(err)
	}
}
