package httransform

import (
	"sync"
	"time"
)

// Tracer allows to trace how HTTP proxy works. On every event happening
// it executes a callback. After request processing is done, it executes
// Dump.
type Tracer interface {
	// StartOnRequest is executed before calling OnRequest for every layer.
	StartOnRequest()

	// StartOnResponse is executed before calling OnResponse for every
	// layer.
	StartOnResponse()

	// StartExecute is executed before calling an executor of the request.
	StartExecute()

	// FinishOnRequest is executed after calling OnRequest for every layer.
	FinishOnRequest(err error)

	// FinishOnResponse is executed after calling OnResponse for every
	// layer.
	FinishOnResponse()

	// FinishExecute is executed after calling executor of the request.
	FinishExecute()

	// Clear drops internal state of the tracer before returning it back to
	// the TracerPool.
	Clear()

	// Dump dumps internal state of the tracer when request is finished its
	// execution.
	Dump(state *LayerState, logger Logger)
}

// NoopTracer is a tracer which does nothing.
type NoopTracer struct {
}

// StartOnRequest is executed before calling OnRequest for every layer.
func (n *NoopTracer) StartOnRequest() {
}

// StartOnResponse is executed before calling OnResponse for every
// layer.
func (n *NoopTracer) StartOnResponse() {
}

// StartExecute is executed before calling an executor of the request.
func (n *NoopTracer) StartExecute() {
}

// FinishOnRequest is executed after calling OnRequest for every layer.
func (n *NoopTracer) FinishOnRequest(_ error) {
}

// FinishOnResponse is executed after calling OnResponse for every
// layer.
func (n *NoopTracer) FinishOnResponse() {
}

// FinishExecute is executed after calling executor of the request.
func (n *NoopTracer) FinishExecute() {
}

// Clear drops internal state of the tracer before returning it back to
// the TracerPool.
func (n *NoopTracer) Clear() {
}

// Dump dumps internal state of the tracer when request is finished its
// execution.
func (n *NoopTracer) Dump(_ *LayerState, _ Logger) {
}

// LogTracer stores duration of execution for OnRequest/OnResponse of
// every layer as well as time elapsed in executor and dumps it to
// logger.
type LogTracer struct {
	startOnRequestTime  time.Time
	startOnResponseTime time.Time
	startExecuteTime    time.Time
	onRequestError      error

	onRequestDurations  []time.Duration
	onResponseDurations []time.Duration
	executeDuration     time.Duration
}

// StartOnRequest is executed before calling OnRequest for every layer.
func (l *LogTracer) StartOnRequest() {
	if !l.startOnRequestTime.IsZero() {
		panic("Start on request already set")
	}
	l.startOnRequestTime = time.Now()
}

// StartOnResponse is executed before calling OnResponse for every
// layer.
func (l *LogTracer) StartOnResponse() {
	if !l.startOnResponseTime.IsZero() {
		panic("Start on response already set")
	}
	l.startOnResponseTime = time.Now()
}

// StartExecute is executed before calling an executor of the request.
func (l *LogTracer) StartExecute() {
	if !l.startExecuteTime.IsZero() {
		panic("Start on execution already set")
	}
	l.startExecuteTime = time.Now()
}

// FinishOnRequest is executed after calling OnRequest for every layer.
func (l *LogTracer) FinishOnRequest(err error) {
	if l.onRequestError != nil {
		panic("OnRequest error already set")
	}
	l.onRequestError = err

	if l.startOnRequestTime.IsZero() {
		panic("Unknown startOnRequest time")
	}

	l.onRequestDurations = append(l.onRequestDurations, time.Since(l.startOnRequestTime))
	l.startOnRequestTime = time.Time{}
}

// FinishOnResponse is executed after calling OnResponse for every
// layer.
func (l *LogTracer) FinishOnResponse() {
	if l.startOnResponseTime.IsZero() {
		panic("Unknown startOnResponseTime time")
	}

	l.onResponseDurations = append(l.onResponseDurations, time.Since(l.startOnResponseTime))
	l.startOnResponseTime = time.Time{}
}

// FinishExecute is executed after calling executor of the request.
func (l *LogTracer) FinishExecute() {
	if l.startExecuteTime.IsZero() {
		panic("Unknown startExecuteTime")
	}

	l.executeDuration = time.Since(l.startExecuteTime)
	l.startExecuteTime = time.Time{}
}

// Clear drops internal state of the tracer before returning it back to
// the TracerPool.
func (l *LogTracer) Clear() {
	l.startOnRequestTime = time.Time{}
	l.startOnResponseTime = time.Time{}
	l.startExecuteTime = time.Time{}
	l.onRequestError = nil
	l.executeDuration = 0

	if len(l.onResponseDurations) > 0 {
		l.onResponseDurations = l.onResponseDurations[:0]
	}
	if len(l.onRequestDurations) > 0 {
		l.onRequestDurations = l.onRequestDurations[:0]
	}
}

// Dump dumps internal state of the tracer when request is finished its
// execution.
func (l *LogTracer) Dump(state *LayerState, logger Logger) {
	responseDurations := make([]time.Duration, len(l.onResponseDurations))
	for i, j := 0, len(l.onResponseDurations)-1; i < j; i, j = i+1, j-1 {
		responseDurations[i], responseDurations[j] = l.onResponseDurations[j], l.onResponseDurations[i]
	}

	logger.Debug("Layer trace",
		"request-id", state.RequestID,
		"on-request-durations", l.onRequestDurations,
		"execute-duration", l.executeDuration,
		"on-response-durations", responseDurations,
		"on-request-error", l.onRequestError,
	)
}

// TracerCreateFunc is a function which creates a new tracer. This
// function has to be used for initialization of the TracerPool.
type TracerCreateFunc func() Tracer

// TracerPool is a special instance of sync.Pool which manages tracer
// objects.
type TracerPool struct {
	sync.Pool
}

func (t *TracerPool) acquire() Tracer {
	return t.Pool.Get().(Tracer)
}

func (t *TracerPool) release(instance Tracer) {
	instance.Clear()
	t.Pool.Put(instance)
}

// NewTracerPool creates a new instance of TracerPool based on the given
// create function.
func NewTracerPool(create TracerCreateFunc) *TracerPool {
	return &TracerPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return create()
			},
		},
	}
}

var defaultNoopTracerPool = NewTracerPool(func() Tracer {
	return &NoopTracer{}
})
