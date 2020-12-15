package upgrades

import "context"

// TCPReactor defines a set of callbacks user has to use to process TCP
// messages.
//
// These callbacks are executed in blocking mode. Please do necessary
// things to prevent corks and bottlenecks there.
type TCPReactor interface {
	// ClientBytes is executed when we WRITE to client socket (netloc ->
	// client).
	ClientBytes(context.Context, []byte)

	// ClientError is executed when we FAILED to write to client socket.
	ClientError(context.Context, error)

	// NetlocBytes is executed when we WRITE to netloc socket (client ->
	// netloc).
	NetlocBytes(context.Context, []byte)

	// NetlocError is executed when we FAILED to write to netloc socket.
	NetlocError(context.Context, error)
}

// NoopTCPReactor is TCP reactor which does nothing.
type NoopTCPReactor struct{}

// ClientBytes conforms TCPReactor interface.
func (n NoopTCPReactor) ClientBytes(_ context.Context, _ []byte) {}

// ClientError conforms TCPReactor interface.
func (n NoopTCPReactor) ClientError(_ context.Context, _ error) {}

// NetlocBytes conforms TCPReactor interface.
func (n NoopTCPReactor) NetlocBytes(_ context.Context, _ []byte) {}

// NetlocError conforms TCPReactor interface.
func (n NoopTCPReactor) NetlocError(_ context.Context, _ error) {}
