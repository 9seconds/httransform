package dialers

import (
	"context"
	"net"
)

type Conn interface {
	net.Conn

	Release()
}

type Dialer interface {
	Dial(context.Context, string) (Conn, error)
	Shutdown()
}
