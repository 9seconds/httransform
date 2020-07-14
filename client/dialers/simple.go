package dialers

import (
	"context"
	"net"
	"time"

	"github.com/PumpkinSeed/errors"
	"github.com/libp2p/go-reuseport"
)

type simpleDialer struct {
	net.Dialer
	baseDialer
}

func (s *simpleDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	conn, err := s.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, ErrDialer)
	}

	return conn, nil
}

func NewSimpleDialer(timeout time.Duration) Dialer {
	return &simpleDialer{
		Dialer: net.Dialer{
			Timeout: timeout,
			Control: reuseport.Control,
		},
		baseDialer: baseDialer{
			timeout: timeout,
		},
	}
}
