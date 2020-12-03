package events

import "context"

// Processor defines an interface for structs which process events.
// These structs are not passed in any function and it is guaranteed
// that messages with the same shardKey are going to be routed to the
// same instance of processor.
//
// It is possible that httransform will create many instances. Each
// instance is going to work independently from each other.
type Processor interface {
	// Process should process an incoming event.
	Process(Event)

	// Shutdown is executed when httransform is going to terminate this
	// processor. No events are going to be passed to this processor
	// once this function is executed.
	Shutdown()
}

// ProcessorFactory defines how to generate new Processor instances.
type ProcessorFactory func() Processor

// Stream defines an interface to event stream.
type Stream interface {
	// Send sends EventType and interface to the stream respecting a
	// given interface and sharding key.
	Send(context.Context, EventType, interface{}, string)
}
