package httransform

import "log"

type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Panic(string, ...interface{})
}

type NoopLogger struct{}

func (n *NoopLogger) Debug(_ string, _ ...interface{}) {}
func (n *NoopLogger) Info(_ string, _ ...interface{})  {}
func (n *NoopLogger) Warn(_ string, _ ...interface{})  {}
func (n *NoopLogger) Error(_ string, _ ...interface{}) {}
func (n *NoopLogger) Panic(_ string, _ ...interface{}) {}

type StdLogger struct {
	Log *log.Logger
}

func (s *StdLogger) Debug(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

func (s *StdLogger) Info(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

func (s *StdLogger) Warn(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

func (s *StdLogger) Error(msg string, args ...interface{}) {
	s.Log.Printf(msg, args...)
}

func (s *StdLogger) Panic(msg string, args ...interface{}) {
	s.Log.Panicf(msg, args...)
}
