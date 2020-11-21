package events

import (
	"context"
	"runtime"
)

var (
	channelShardNumber       = runtime.NumCPU()
	channelShardNumberUint64 = uint64(channelShardNumber)
)

func NewEventChannel(ctx context.Context, processor EventProcessor) EventChannel {
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
			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-eventChannel:
					processor(evt)
				}
			}
		}(shards[i])
	}

	return multiplexChannel
}
