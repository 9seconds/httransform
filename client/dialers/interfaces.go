package dialers

import (
	"context"
	"net"
	"time"
)

type Dialer interface {
	Dial(context.Context, string) (net.Conn, error)
	GetTimeout() time.Duration
}
