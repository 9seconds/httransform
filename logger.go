package httransform

import (
	"fmt"
	"log"
)

// Logger is the common interface for the logger which can be used
// within httransform internals. Basically, if you need to log something
// from internals, please propagate something which implements this
// interface to NewServer.
//
// All methods of this interface should work in Print mode: messages and
// arguments.
type Logger interface {
	// Debug logs information which is usable only for debugging. Usually
	// this is quite noisy data you absolutely do not want to have in
	// production.
	Debug(string, ...interface{})

	// Info logs some informational messages about the current state.
	Info(string, ...interface{})

	// Warn logs events which are suspicious but the system can continue to
	// work.
	Warn(string, ...interface{})

	// Error logs events which are serious problems which can corrupt your
	// server.
	Error(string, ...interface{})

	// Panic logs events which lead to crash.
	Panic(string, ...interface{})
}

// NoopLogger is a dummy logger which drops out messages.
type NoopLogger struct{}

// Debug is to support Logger interface.
func (n *NoopLogger) Debug(_ string, _ ...interface{}) {}

// Info is to support Logger interface.
func (n *NoopLogger) Info(_ string, _ ...interface{}) {}

// Warn is to support Logger interface.
func (n *NoopLogger) Warn(_ string, _ ...interface{}) {}

// Error is to support Logger interface.
func (n *NoopLogger) Error(_ string, _ ...interface{}) {}

// Panic is to support Logger interface.
func (n *NoopLogger) Panic(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}

// StdLogger is the wrapper over default log.Logger instance which provides
// Logger inteface for httransform.
type StdLogger struct {
	// Log is the configured instance of logger we want to wrap.
	Log *log.Logger
}

// Debug is to support Logger interface.
func (s *StdLogger) Debug(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

// Info is to support Logger interface.
func (s *StdLogger) Info(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

// Warn is to support Logger interface.
func (s *StdLogger) Warn(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

// Error is to support Logger interface.
func (s *StdLogger) Error(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

// Panic is to support Logger interface.
func (s *StdLogger) Panic(msg string, args ...interface{}) {
	s.Log.Panicf(msg, args...)
}
