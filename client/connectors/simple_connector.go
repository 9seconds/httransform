package connectors

import "context"

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

func NewSimpleConnector(ctx context.Context, dialer Dialer) Connector {
	ctx, cancel := context.WithCancel(ctx)

	return &simpleConnector{
		baseConnector: baseConnector{
			dialer: dialer,
			ctx:    ctx,
			cancel: cancel,
		},
	}
}
