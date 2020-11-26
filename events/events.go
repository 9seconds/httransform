package events

import (
	"math/rand"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
)

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

	var shard int

	if shardKey == "" {
		shard = rand.Intn(channelShardNumber) // nolint: gosec
	} else {
		shard = int(xxhash.ChecksumString64(shardKey) % channelShardNumberUint64)
	}

	evt.id = eventType
	evt.timestamp = time.Now()
	evt.value = value
	evt.shard = shard

	return evt
}

func ReleaseEvent(evt *Event) {
	evt.Reset()
	poolEvent.Put(evt)
}
