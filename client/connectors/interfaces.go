package connectors

import (
	"context"
	"net"
)

type Conn interface {
	net.Conn

	Release()
}

type Connector interface {
	Connect(context.Context, string) (Conn, error)
	Shutdown()
}
