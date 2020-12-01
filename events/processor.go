package events

type noopProcessor struct{}

func (n noopProcessor) Process(_ *Event) {}

func NoopProcessorFactory() Processor {
	return noopProcessor{}
}
