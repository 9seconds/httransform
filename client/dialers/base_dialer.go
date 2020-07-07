package dialers

import (
	"context"
	"math/rand"
	"net"

	"github.com/PumpkinSeed/errors"
)

type baseDialer struct {
	dialer net.Dialer
	ctx    context.Context
	cancel context.CancelFunc
}

func (b *baseDialer) dial(ctx context.Context, addr string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, b.dialer.Timeout)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
		case <-b.ctx.Done():
			cancel()
		}
	}()

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, errors.Wrap(err, ErrCannotSplitHostPort)
	}

	ips, err := dns.LookupHost(ctx, host)
	if err != nil {
		return nil, errors.Wrap(err, ErrDNSError)
	}

	if len(ips) == 0 {
		return nil, ErrNoIPs
	}

	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})

	var conn net.Conn

	for _, ip := range ips {
		conn, err = b.dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))

		if err == nil {
			return conn, nil
		}

		if conn != nil {
			conn.Close()
		}
	}

	return nil, errors.Wrap(err, ErrCannotDial)
}

func (b *baseDialer) Shutdown() {
	b.cancel()
}
