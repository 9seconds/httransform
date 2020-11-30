package upgrades

import "context"

type TCPReactor interface {
	ClientBytes(context.Context, []byte)
	ClientError(context.Context, error)
	NetlocBytes(context.Context, []byte)
	NetlocError(context.Context, error)
}

type NoopTCPReactor struct{}

func (n NoopTCPReactor) ClientBytes(_ context.Context, b []byte) {}
func (n NoopTCPReactor) ClientError(_ context.Context, _ error)  {}
func (n NoopTCPReactor) NetlocBytes(_ context.Context, b []byte) {}
func (n NoopTCPReactor) NetlocError(_ context.Context, _ error)  {}
