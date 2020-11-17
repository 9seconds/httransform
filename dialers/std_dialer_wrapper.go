package dialers

import (
	"context"
	"fmt"
	"net"
)

type StdDialerWrapper struct {
	wrapped Dialer
}

func (s StdDialerWrapper) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" {
		return nil, fmt.Errorf("Incorrect network %s", network)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("incorrect address: %w", err)
	}

	return s.wrapped.Dial(ctx, host, port)
}

func (s StdDialerWrapper) Dial(network, address string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, address)
}
