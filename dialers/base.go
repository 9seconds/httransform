package dialers

import (
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/9seconds/httransform/v2/dns"
	"github.com/9seconds/httransform/v2/errors"
	"github.com/libp2p/go-reuseport"
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
	tlsConfigsLock sync.Mutex
	tlsConfigs     cache.Interface
	tlsSkipVerify  bool
}

func (b *base) Dial(ctx context.Context, host, port string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, b.netDialer.Timeout)
	defer cancel()

	ips, err := dns.Default.Lookup(ctx, host)
	if err != nil {
		return nil, errors.Annotate(err, "cannot resolve IPs", "dns_no_ips", 0)
	}

	if len(ips) == 0 {
		return nil, ErrNoIPs
	}

	var conn net.Conn

	for _, ip := range ips {
		conn, err = b.netDialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
		if err == nil {
			return conn, nil
		}
	}

	return nil, errors.Annotate(err, "cannot dial to "+host, "cannot_dial", 0)
}

func (b *base) UpgradeToTLS(ctx context.Context, conn net.Conn, host, _ string) (net.Conn, error) {
	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		timer := fasthttp.AcquireTimer(b.netDialer.Timeout)
		defer fasthttp.ReleaseTimer(timer)

		select {
		case <-subCtx.Done():
		case <-ctx.Done():
			select {
			case <-subCtx.Done():
			default:
				conn.Close()
			}
		case <-timer.C:
			select {
			case <-subCtx.Done():
			default:
				conn.Close()
			}
		}
	}()

	tlsConn := tls.Client(conn, b.getTLSConfig(host))
	if err := tlsConn.Handshake(); err != nil {
		return nil, errors.Annotate(err, "cannot perform TLS handshake", "tls_handshake", 0)
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

	return rv
}
