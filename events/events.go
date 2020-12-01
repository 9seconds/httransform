package events

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/OneOfOne/xxhash"
)

var eventCounter uint64

type EventType byte

const (
	EventTypeNotSet          EventType = iota // fake value
	EventTypeNewCertificate                   // value - hostname(string)
	EventTypeDropCertificate                  // value - hostname(string)
	EventTypeFailedAuth                       // value - nil
	EventTypeStartRequest                     // value - requestMeta
	EventTypeFailedRequest                    // value - errorMeta
	EventTypeFinishRequest                    // value - responseMeta
	EventTypeTraffic                          // value - trafficMeta
	EventTypeUserValues                       // fake value
)

type Event struct {
	id        EventType
	timestamp time.Time
	shard     int
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

func AcquireEvent(eventType EventType, value interface{}, shardKey string) *Event {
	evt := poolEvent.Get().(*Event)

	var shardValue uint64

	if shardKey == "" {
		shardValue = atomic.AddUint64(&eventCounter, 1)
	} else {
		shardValue = xxhash.Checksum64([]byte(shardKey))
	}

	evt.id = eventType
	evt.timestamp = time.Now()
	evt.value = value
	evt.shard = int(shardValue % channelShardNumberUint64)

	return evt
}

func ReleaseEvent(evt *Event) {
	evt.Reset()
	poolEvent.Put(evt)
}
