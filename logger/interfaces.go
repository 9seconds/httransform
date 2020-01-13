package logger

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
