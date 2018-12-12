package httransform

import "time"

type tracer interface {
	startOnRequest()
	startOnResponse()
	startExecute()
	finishOnRequest(error)
	finishOnResponse()
	finishExecute()
	clear()

	dump(uint64, Logger)
}

type dummyTracer struct {
}

func (d *dummyTracer) startOnRequest() {
}

func (d *dummyTracer) startOnResponse() {
}

func (d *dummyTracer) startExecute() {
}

func (d *dummyTracer) finishOnRequest(_ error) {
}

func (d *dummyTracer) finishOnResponse() {
}

func (d *dummyTracer) finishExecute() {
}

func (d *dummyTracer) clear() {
}

func (d *dummyTracer) dump(_ uint64, _ Logger) {
}

type logTracer struct {
	startOnRequestTime  time.Time
	startOnResponseTime time.Time
	startExecuteTime    time.Time
	onRequestError      error

	onRequestDurations  []time.Duration
	onResponseDurations []time.Duration
	executeDuration     time.Duration
}

func (l *logTracer) startOnRequest() {
	if !l.startOnRequestTime.IsZero() {
		panic("Start on request already set")
	}
	l.startOnRequestTime = time.Now()
}

func (l *logTracer) startOnResponse() {
	if !l.startOnResponseTime.IsZero() {
		panic("Start on response already set")
	}
	l.startOnResponseTime = time.Now()
}

func (l *logTracer) startExecute() {
	if !l.startExecuteTime.IsZero() {
		panic("Start on execution already set")
	}
	l.startExecuteTime = time.Now()
}

func (l *logTracer) finishOnRequest(err error) {
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

func (l *logTracer) finishOnResponse() {
	if l.startOnResponseTime.IsZero() {
		panic("Unknown startOnResponseTime time")
	}

	l.onResponseDurations = append(l.onResponseDurations, time.Since(l.startOnResponseTime))
	l.startOnResponseTime = time.Time{}
}

func (l *logTracer) finishExecute() {
	if l.startExecuteTime.IsZero() {
		panic("Unknown startExecuteTime")
	}

	l.executeDuration = time.Since(l.startExecuteTime)
}

func (l *logTracer) clear() {
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

func (l *logTracer) dump(requestID uint64, logger Logger) {
	responseDurations := make([]time.Duration, len(l.onResponseDurations))
	for i, j := 0, len(l.onResponseDurations)-1; i < j; i, j = i+1, j-1 {
		responseDurations[i], responseDurations[j] = l.onResponseDurations[j], l.onResponseDurations[i]
	}

	logger.Debug("Layer trace",
		"request-id", requestID,
		"on-request-durations", l.onRequestDurations,
		"execute-duration", l.executeDuration,
		"on-response-durations", responseDurations,
		"on-request-error", l.onRequestError,
	)
}
