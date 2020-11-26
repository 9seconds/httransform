package events

import "context"

type EventProcessor func(*Event)

type EventChannel chan<- *Event

func (e EventChannel) Send(ctx context.Context, eventType EventType, value interface{}, shardKey string) {
    evt := AcquireEvent(eventType, value, shardKey)

    select {
    case <-ctx.Done():
        ReleaseEvent(evt)
    case e <- evt:
    }
}
