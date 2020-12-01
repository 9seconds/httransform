package events

type Processor interface {
	Process(*Event)
}

type ProcessorFactory func() Processor
