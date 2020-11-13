package dialer

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/libp2p/go-reuseport"
	"github.com/rs/dnscache"
)

type base struct {
	net.Dialer

	dnsCache dnscache.Resolver
}

func (b *base) Dial(network, address string) (net.Conn, error) {
	return b.DialContext(context.Background(), network, address)
}

func (b *base) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, b.Timeout)
	defer cancel()

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("cannot dial: %w", err)
	}

	ips, err := b.dnsCache.LookupHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("cannot reslolve IPs: %w", err)
	}

	if len(ips) == 0 {
		return nil, ErrNoIPs
	}

	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})

	var conn net.Conn

	for _, ip := range ips {
		conn, err = b.Dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
		if err == nil {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("cannot dial to %s: %w", host, err)
}

func NewBase(ctx context.Context, timeout time.Duration) Dialer {
	rv := &base{
		Dialer: net.Dialer{
			Timeout: timeout,
			Control: reuseport.Control,
		},
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rv.dnsCache.Refresh(true)
			}
		}
	}()

	return rv
}
