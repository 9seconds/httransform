package events

import (
	"sync"
	"time"
)

type EventType byte

const (
	EventTypeNotSet          EventType = iota // fake value
	EventTypeNewCertificate                   // value - hostname(string)
	EventTypeDropCertificate                  // value - hostname(string)
	EventTypeFailedAuth                       // value - nil
	EventTypeStartRequest                     // value - requestMeta
	EventTypeFinishRequest                    // value - responseMeta
	EventTypeUserValues                       // fake value
)

type Event struct {
	id        EventType
	timestamp time.Time
	value     interface{}
}

func (e *Event) ID() EventType {
	return e.id
}

func (e *Event) TimeStamp() time.Time {
	return e.timestamp
}

func (e *Event) Value() interface{} {
	return e.value
}

func (e *Event) Reset() {
	*e = Event{
		id: EventTypeNotSet,
	}
}

var poolEvent = sync.Pool{
	New: func() interface{} {
		return &Event{
			id: EventTypeNotSet,
		}
	},
}

func AcquireEvent(eventType EventType, value interface{}) *Event {
	evt := poolEvent.Get().(*Event)

	evt.id = eventType
	evt.timestamp = time.Now()
	evt.value = value

	return evt
}

func ReleaseEvent(evt *Event) {
	evt.Reset()
	poolEvent.Put(evt)
}
