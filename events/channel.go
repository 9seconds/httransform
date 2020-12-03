package events

import (
	"context"
	"runtime"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/valyala/fastrand"
)

// Channel is just a wrapper over a chan *Event but since we do pooling
// of events under the hood, it is hidden.
//
// When you work with a Channel instance, you work with a single chan
// But at the same time, it does some 'sharding' and route a message
// based on a shardKey in a different channel. Each end-channel is
// managed by a single instance of processor. So, it is guaranteed that
// a message with the same shardKey is going to be routed to the same
// instance of Processor. If shardKey is an empty string, a random
// channel is picked.
type Channel struct {
	ctx    context.Context
	shards []chan Event
}

// Send sends an event to a processor.
//
// If given ctx is cancelled, this function exits. shardKey is necessary
// to establish a permanent routing so messages with the same shardKey
// will be send to the same processor instance.
func (c *Channel) Send(ctx context.Context, eventType EventType, value interface{}, shardKey string) {
	var shard int

	if shardKey == "" {
		shard = int(fastrand.Uint32() % uint32(len(c.shards)))
	} else {
		shard = int(xxhash.ChecksumString64(shardKey) % uint64(len(c.shards)))
	}

	evt := Event{
		Type:  eventType,
		Time:  time.Now(),
		Value: value,
	}

	select {
	case <-ctx.Done():
	case <-c.ctx.Done():
	case c.shards[shard] <- evt:
	}
}

// NewChannel creates, initializes and returns a ready Channel instance.
// It spawns a set of worker goroutines under the hood. Each goroutine
// corresponds to a its own processor instance (that's why you pass
// factory here). Processor is initialized within a goroutine.
func NewChannel(ctx context.Context, factory ProcessorFactory) *Channel {
	rv := &Channel{
		ctx:    ctx,
		shards: make([]chan Event, runtime.NumCPU()),
	}

	for i := range rv.shards {
		rv.shards[i] = make(chan Event, 1)
	}

	for _, v := range rv.shards {
		go func(channel <-chan Event) {
			processor := factory()
			defer processor.Shutdown()

			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-channel:
					processor.Process(evt)
				}
			}
		}(v)
	}

	return rv
}
