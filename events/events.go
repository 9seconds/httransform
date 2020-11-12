package events

import (
	"sync"
	"time"
)

type EventType byte

const (
	EventTypeNewCertificate  EventType = iota // value - hostname(string)
	EventTypeDropCertificate                  // value - hostname(string)
)

type Event struct {
	ID        EventType
	Timestamp time.Time
	Value     interface{}
}

type EventProcessor func(*Event)

var poolEvents = sync.Pool{
	New: func() interface{} {
		return Event{}
	},
}

func acquireEvent() *Event {
	return poolEvents.Get().(*Event)
}

func releaseEvent(evt *Event) {
	evt.Value = nil
	poolEvents.Put(evt)
}
