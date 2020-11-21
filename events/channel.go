package events

import (
	"context"
	"runtime"
)

func NewEventChannel(ctx context.Context, processor EventProcessor) EventChannel {
	numCPU := runtime.NumCPU()
	eventChan := make(chan *Event, numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-eventChan:
					processor(evt)
					ReleaseEvent(evt)
				}
			}
		}()
	}

	return eventChan
}
