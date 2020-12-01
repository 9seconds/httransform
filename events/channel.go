package events

import (
	"context"
	"runtime"
)

var (
	channelShardNumber       = runtime.NumCPU()
	channelShardNumberUint64 = uint64(channelShardNumber)
)

type Channel chan<- *Event

func (c Channel) Send(ctx context.Context, eventType EventType, value interface{}, shardKey string) {
	evt := acquireEvent(eventType, value, shardKey)

	select {
	case <-ctx.Done():
		releaseEvent(evt)
	case c <- evt:
	}
}

func NewChannel(ctx context.Context, factory ProcessorFactory) Channel {
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

	return multiplexChannel
}
