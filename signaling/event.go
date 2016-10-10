package signaling

import "encoding/json"

// Kinder ...
type Kinder interface {
	Kind() string
}

var register = map[string]func() Kinder{}

// Register ...
func Register(factory func() Kinder) {
	obj := factory()
	register[obj.Kind()] = factory
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

// Join ...
type Join struct {
	Member string
}

// Kind ...
func (c *Join) Kind() string { return "join" }

// Leave ...
type Leave struct {
	Member string
}

// Kind ...
func (c *Leave) Kind() string { return "leave" }

func init() {
	Register(func() Kinder { return new(Join) })
	Register(func() Kinder { return new(Leave) })
}
