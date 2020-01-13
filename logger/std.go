package logger

import "log"

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
