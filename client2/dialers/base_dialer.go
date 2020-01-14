package dialers

import (
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/rs/dnscache"
)

const BaseDialerDNSRefreshEvery = time.Minute

type baseDialer struct {
	dialer net.Dialer
	dns    dnscache.Resolver
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
		return nil, ErrBaseDialerCannotParse.Wrap("cannot parse address", err)
	}

	ips, err := b.dns.LookupHost(ctx, host)
	if err != nil {
		return nil, ErrBaseDialerDNS.Wrap("dns lookup failed", err)
	}

	if len(ips) == 0 {
		return nil, ErrBaseDialerNoIPs
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

	return nil, ErrBaseDialerAllFailed.Wrap("all ips failed", err)
}

func (b *baseDialer) Shutdown() {
	b.cancel()
}

func (b *baseDialer) run() {
	ticker := time.NewTicker(BaseDialerDNSRefreshEvery)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.dns.Refresh(true)
		}
	}
}
