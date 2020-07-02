package dialers

import (
	"context"
	"net"
	"time"

	reuseport "github.com/libp2p/go-reuseport"
)

const SimpleDialerTimeout = 30 * time.Second

type simpleDialer struct {
	baseDialer
}

func (s *simpleDialer) Dial(ctx context.Context, addr string) (Conn, error) {
	conn, err := s.dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &simpleConn{TCPConn: *conn.(*net.TCPConn)}, nil
}

func NewSimpleDialer(ctx context.Context, timeout time.Duration) Dialer {
	if timeout == 0 {
		timeout = SimpleDialerTimeout
	}

	ctx, cancel := context.WithCancel(ctx)
	rv := &simpleDialer{
		baseDialer: baseDialer{
			dialer: net.Dialer{
				Timeout: timeout,
				Control: reuseport.Control,
			},
			ctx:    ctx,
			cancel: cancel,
		},
	}

	return rv
}
