package events

import (
	"context"
	"time"
)

type EventStream struct {
	ctx           context.Context
	channelEvents chan *Event
	processor     EventProcessor
}

func (e *EventStream) Send(eventType EventType, value interface{}) {
	evt := acquireEvent()
	evt.ID = eventType
	evt.Value = value
	evt.Timestamp = time.Now()

	select {
	case <-e.ctx.Done():
	case e.channelEvents <- evt:
	}
}

func NewEventStream(ctx context.Context, processor EventProcessor) *EventStream {
	rv := &EventStream{
		ctx:           ctx,
		channelEvents: make(chan *Event, 1),
		processor:     processor,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-rv.channelEvents:
				rv.processor(evt)
				releaseEvent(evt)
			}
		}
	}()

	return rv
}
