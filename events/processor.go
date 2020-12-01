package events

type noopProcessor struct{}

func (n noopProcessor) Process(_ *Event) {}
func (n noopProcessor) Shutdown()        {}

func NoopProcessorFactory() Processor {
	return noopProcessor{}
}
