package events

// Processor defines an interface for structs which process events.
// These structs are not passed in any function and it is guaranteed
// that messages with the same shardKey are going to be routed to the
// same instance of processor.
//
// It is possible that httransform will create many instances. Each
// instance is going to work independently from each other.
type Processor interface {
	// Process should process an incoming event. It is crucial that
	// nothing should refer to an event once processing is finished.
	//
	// All events are pooled so referencing them after processing can
	// lead to undefined behaviour.
	Process(*Event)

	// Shutdown is executed when httransform is going to terminate this
	// processor. No events are going to be passed to this processor
	// once this function is executed.
	Shutdown()
}

// ProcessorFactory defines how to generate new Processor instances.
type ProcessorFactory func() Processor
