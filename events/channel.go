package events

import (
	"context"
	"runtime"
)

func NewEventChannel(ctx context.Context, processor EventProcessor) chan<- Event {
	numCPU := runtime.NumCPU()
	eventChan := make(chan Event, numCPU)

	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-eventChan:
					processor(evt)
				}
			}
		}()
	}

	return eventChan
}
