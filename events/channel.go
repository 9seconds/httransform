package events

import (
	"context"
	"runtime"
)

var (
	channelShardNumber       = runtime.NumCPU()
	channelShardNumberUint64 = uint64(channelShardNumber)
)

// Channel is just a wrapper over a chan *Event but since we do pooling
// of events under the hood, it is hidden.
//
// When you work with a Channel instance, you work with a single chan
// But at the same time, it does some 'sharding' and route a message
// based on a shardKey in a different channel. Each end-channel is
// managed by a single instance of processor. So, it is guaranteed that
// a message with the same shardKey is going to be routed to the same
// instance of Processor. If shardKey is an empty string, a channel is
// selected in round-robin fashion.
type Channel struct {
	ctx    context.Context
	events chan<- *Event
}

// Send sends an event to a processor.
//
// If given ctx is cancelled, this function exits. shardKey is necessary
// to establish a permanent routing so messages with the same shardKey
// will be send to the same processor instance.
func (c *Channel) Send(ctx context.Context, eventType EventType, value interface{}, shardKey string) {
	evt := acquireEvent(eventType, value, shardKey)

	select {
	case <-ctx.Done():
		releaseEvent(evt)
	case <-c.ctx.Done():
		releaseEvent(evt)
	case c.events <- evt:
	}
}

// NewChannel creates, initializes and returns a ready Channel instance.
// It spawns a set of worker goroutines under the hood. Each goroutine
// corresponds to a its own processor instance (that's why you pass
// factory here). Processor is initialized within a goroutine.
func NewChannel(ctx context.Context, factory ProcessorFactory) *Channel {
	multiplexChannel := make(chan *Event, channelShardNumber)
	shards := make([]chan *Event, channelShardNumber)

	for i := range shards {
		shards[i] = make(chan *Event, 1)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-multiplexChannel:
				shards[evt.shard] <- evt
			}
		}
	}()

	for i := range shards {
		go func(eventChannel <-chan *Event) {
			processor := factory()
			defer processor.Shutdown()

			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-eventChannel:
					processor.Process(evt)
					releaseEvent(evt)
				}
			}
		}(shards[i])
	}

	return &Channel{
		ctx:    ctx,
		events: multiplexChannel,
	}
}
