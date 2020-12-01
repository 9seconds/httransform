package events

type Processor interface {
	Process(*Event)
	Shutdown()
}

type ProcessorFactory func() Processor
