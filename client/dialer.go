package client

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/rs/dnscache"
)

type Dialer struct {
	net.Dialer
	dnsCache dnscache.Resolver
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("cannot dial: %w", err)
	}

	ips, err := d.dnsCache.LookupHost(ctx, host)
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
		conn, err = d.Dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
		if err == nil {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("cannot dial to %s: %w", host, err)
}
