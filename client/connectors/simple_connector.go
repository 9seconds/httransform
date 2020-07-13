package connectors

import (
	"context"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
)

type simpleConnector struct {
	baseConnector
}

func (s *simpleConnector) Connect(ctx context.Context, addr string) (Conn, error) {
	conn, err := s.connect(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &simpleConn{
		Conn: conn,
	}, nil
}

func NewSimpleConnector(ctx context.Context, timeout time.Duration) Connector {
	if timeout == 0 {
		timeout = SimpleConnectorTimeout
	}

	ctx, cancel := context.WithCancel(ctx)

	return &simpleConnector{
		baseConnector: baseConnector{
			dialer: net.Dialer{
				Timeout: timeout,
				Control: reuseport.Control,
			},
			ctx:    ctx,
			cancel: cancel,
		},
	}
}
