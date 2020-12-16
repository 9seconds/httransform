package events

type noopProcessor struct{}

func (n noopProcessor) Process(_ Event) {}
func (n noopProcessor) Shutdown()       {}

// NoopProcessorFactory returns a processor which does nothing and skips
// all events.
func NoopProcessorFactory() Processor {
	return noopProcessor{}
}
