package dialers

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/libp2p/go-reuseport"
	"github.com/rs/dnscache"
)

type base struct {
	netDialer      net.Dialer
	dnsCache       dnscache.Resolver
	tlsConfigsLock sync.Mutex
	tlsConfigs     *lru.TwoQueueCache
	tlsSkipVerify  bool
}

func (b *base) Dial(ctx context.Context, host, port string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, b.netDialer.Timeout)
	defer cancel()

	ips, err := b.dnsCache.LookupHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve IPs: %w", err)
	}

	if len(ips) == 0 {
		return nil, ErrNoIPs
	}

	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})

	var conn net.Conn

	for _, ip := range ips {
		conn, err = b.netDialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
		if err == nil {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("cannot dial to %s: %w", host, err)
}

func (b *base) UpgradeToTLS(ctx context.Context, conn net.Conn, host string) (net.Conn, error) {
	ownCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-ownCtx.Done():
		case <-ctx.Done():
			conn.Close()
		}
	}()

	tlsConn := tls.Client(conn, b.getTLSConfig(host))
	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("cannot perform TLS handshake: %w", err)
	}

	return tlsConn, nil
}

func (b *base) getTLSConfig(host string) *tls.Config {
	if conf, ok := b.tlsConfigs.Get(host); ok {
		return conf.(*tls.Config)
	}

	b.tlsConfigsLock.Lock()
	defer b.tlsConfigsLock.Unlock()

	if conf, ok := b.tlsConfigs.Get(host); ok {
		return conf.(*tls.Config)
	}

	conf := &tls.Config{
		ClientSessionCache: tls.NewLRUClientSessionCache(0),
		ServerName:         host,
		InsecureSkipVerify: b.tlsSkipVerify,
	}

	b.tlsConfigs.Add(host, conf)

	return conf
}

func NewBase(opt Opts) (Dialer, error) {
	tlsConfigs, err := lru.New2Q(opt.GetTLSConfigCacheMaxSize())
	if err != nil {
		return nil, fmt.Errorf("cannot build lru cache for tls configs: %w", err)
	}

	rv := &base{
		netDialer: net.Dialer{
			Timeout: opt.GetTimeout(),
			Control: reuseport.Control,
		},
		tlsConfigs:    tlsConfigs,
		tlsSkipVerify: opt.GetTLSSkipVerify(),
	}

	go func(ctx context.Context, period time.Duration) {
		ticker := time.NewTicker(period)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rv.dnsCache.Refresh(true)
			}
		}
	}(opt.GetContext(), opt.GetCleanupDNSEvery())

	return rv, nil
}
