package dialers

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/libp2p/go-reuseport"
	"github.com/rs/dnscache"
	"github.com/valyala/fasthttp"
)

const (
	// TLSConfigCacheSize defines a size of LRU cache which is used by base
	// dialer.
	TLSConfigCacheSize = 512

	// TLSConfigTTL defines a TTL for each tls.Config we generate.
	TLSConfigTTL = 10 * time.Minute
)

type base struct {
	netDialer      net.Dialer
	dnsCache       dnscache.Resolver
	tlsConfigsLock sync.Mutex
	tlsConfigs     cache.Interface
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
	ownCtx, cancel := context.WithTimeout(ctx, b.netDialer.Timeout)
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

func (b *base) PatchHTTPRequest(req *fasthttp.Request) {
	if bytes.EqualFold(req.URI().Scheme(), []byte("http")) {
		req.SetRequestURIBytes(req.URI().PathOriginal())
	}
}

func (b *base) getTLSConfig(host string) *tls.Config {
	if conf := b.tlsConfigs.Get(host); conf != nil {
		return conf.(*tls.Config)
	}

	b.tlsConfigsLock.Lock()
	defer b.tlsConfigsLock.Unlock()

	if conf := b.tlsConfigs.Get(host); conf != nil {
		return conf.(*tls.Config)
	}

	conf := &tls.Config{
		ClientSessionCache: tls.NewLRUClientSessionCache(0),
		ServerName:         host,
		InsecureSkipVerify: b.tlsSkipVerify, // nolint: gosec
	}

	b.tlsConfigs.Add(host, conf)

	return conf
}

// NewBase returns a base dialer which connects to a target website and
// does only those operations which are required:
//
// 1. Dial establishes a TCP connection to a target netloc
//
// 2. UpgradeToTLS upgrades this TCP connection to secured one.
//
// 3. PatchHTTPRequest does processing which makes sense only to adjust
// with fasthttp specific logic.
//
// Apart from that, it sets timeouts, uses SO_REUSEADDR socket option,
// uses DNS cache and reuses tls.Config instances when possible.
func NewBase(opt Opts) Dialer {
	rv := &base{
		netDialer: net.Dialer{
			Timeout: opt.GetTimeout(),
			Control: reuseport.Control,
		},
		tlsConfigs: cache.New(TLSConfigCacheSize,
			TLSConfigTTL,
			cache.NoopEvictCallback),
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
	}(opt.GetContext(), opt.GetDNSUpdateEvery())

	return rv
}
