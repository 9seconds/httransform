package dialers

import (
	"context"
	"fmt"
	"net"
)

// StdDialerWrapper just wraps a Dialer of this package into net.Dialer
// interface.
type StdDialerWrapper struct {
	Dialer Dialer
}

// DialContext is to conform DialContext method of net.Dialer.
func (s StdDialerWrapper) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("incorrect network %s", network)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("incorrect address: %w", err)
	}

	return s.Dialer.Dial(ctx, host, port) // nolint: wrapcheck
}

// Dial is to conform Dial method of net.Dialer.
func (s StdDialerWrapper) Dial(network, address string) (net.Conn, error) {
	return s.DialContext(context.Background(), network, address)
}
