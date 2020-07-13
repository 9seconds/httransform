package connectors

import (
	"context"
	"net"
	"time"
)

type Conn interface {
	net.Conn

	Release()
}

type Connector interface {
	Connect(context.Context, string) (Conn, error)
	Shutdown()
}

type Dialer interface {
	Dial(context.Context, string, string) (net.Conn, error)
	Timeout() time.Duration
}
