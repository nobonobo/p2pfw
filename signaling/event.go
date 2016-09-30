package signaling

import (
	"encoding/json"
	"log"
)

// Kinder ...
type Kinder interface {
	Kind() string
}

var register = map[string]func() Kinder{}

// Register ...
func Register(factory func() Kinder) {
	obj := factory()
	register[obj.Kind()] = factory
	log.Println("event register:", obj.Kind())
}

// Event ...
type Event struct {
	From  string          `json:"from"`
	To    string          `json:"to"`
	Kind  string          `json:"kind"`
	Value json.RawMessage `json:"value"`
}

// New ...
func New(from, to string, value Kinder) *Event {
	ev := &Event{
		From: from,
		To:   to,
		Kind: value.Kind(),
	}
	ev.Value, _ = json.Marshal(value)
	return ev
}

// Get ...
func (ev *Event) Get() Kinder {
	var value Kinder
	if f, ok := register[ev.Kind]; ok {
		value = f()
	}
	if err := json.Unmarshal(ev.Value, &value); err != nil {
		return nil
	}
	return value
}
