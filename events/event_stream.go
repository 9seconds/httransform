package events

import (
	"context"
	"runtime"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/valyala/fastrand"
)

type eventStream struct {
	ctx    context.Context
	shards []chan Event
}

func (e *eventStream) Send(ctx context.Context, eventType EventType, value interface{}, shardKey string) {
	var shard int

	if shardKey == "" {
        shard = int(fastrand.Uint32n(uint32(len(e.shards))))
	} else {
		shard = int(xxhash.ChecksumString64(shardKey) % uint64(len(e.shards)))
	}

	evt := Event{
		Type:  eventType,
		Time:  time.Now(),
		Value: value,
	}

	select {
	case <-ctx.Done():
	case <-e.ctx.Done():
	case e.shards[shard] <- evt:
	}
}

// NewStream creates, initialized and returns a new ready Stream
// instance. It spawns a set of worker goroutines under the hood. Each
// goroutine corresponds to a its own processor instance (that's why you
// pass factory here). Processor is initialized within a goroutine.
func NewStream(ctx context.Context, factory ProcessorFactory) Stream {
	rv := &eventStream{
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
