package dialer

import (
	"context"
	"net"
)

type Dialer interface {
	Dial(string, string) (net.Conn, error)
	DialContext(context.Context, string, string) (net.Conn, error)
}
