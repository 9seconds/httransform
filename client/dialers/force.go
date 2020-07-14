package dialers

import (
	"context"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
)

type forceAddressDialer struct {
	simpleDialer

	address string
}

func (f *forceAddressDialer) Dial(ctx context.Context, _ string) (net.Conn, error) {
	return f.simpleDialer.Dial(ctx, f.address)
}

func NewForceAddressDialer(timeout time.Duration, address string) Dialer {
	return &forceAddressDialer{
		simpleDialer: simpleDialer{
			Dialer: net.Dialer{
				Timeout: timeout,
				Control: reuseport.Control,
			},
			baseDialer: baseDialer{
				timeout: timeout,
			},
		},
		address: address,
	}
}
