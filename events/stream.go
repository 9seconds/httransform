package events

import "context"

type EventStream struct {
	Chan chan Event

	ctx       context.Context
	processor EventProcessor
}

func NewEventStream(ctx context.Context, processor EventProcessor) *EventStream {
	rv := &EventStream{
		ctx:       ctx,
		processor: processor,
		Chan:      make(chan Event, 1),
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-rv.Chan:
				rv.processor(evt)
			}
		}
	}()

	return rv
}
