package events

import "time"

type EventType byte

const (
	EventTypeNewCertificate  EventType = iota // value - hostname(string)
	EventTypeDropCertificate                  // value - hostname(string)
	EventTypeFailedAuth                       // value - nil
	EventTypeStartRequest                     // value - requestMeta
	EventTypeFinishRequest                    // value - responseMeta
)

type Event struct {
	ID        EventType
	Timestamp time.Time
	Value     interface{}
}

type EventProcessor func(Event)

func DefaultProcessor(_ Event) {}

func New(id EventType, value interface{}) Event {
	return Event{
		ID:        id,
		Value:     value,
		Timestamp: time.Now(),
	}
}
