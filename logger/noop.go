package logger

import "fmt"

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
