package dialers

import (
	"context"
	"fmt"
	"net"

	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
)

type socksProxy struct {
	baseDialer *base
	proxy      proxy.ContextDialer
}

func (s *socksProxy) Dial(ctx context.Context, host, port string) (net.Conn, error) {
	return s.proxy.DialContext(ctx, "tcp", net.JoinHostPort(host, port)) // nolint: wrapcheck
}

func (s *socksProxy) UpgradeToTLS(ctx context.Context, conn net.Conn, host, port string) (net.Conn, error) {
	return s.baseDialer.UpgradeToTLS(ctx, conn, host, port)
}

func (s *socksProxy) PatchHTTPRequest(req *fasthttp.Request) {
	s.baseDialer.PatchHTTPRequest(req)
}

// NewSocks5 returns a dialer which can connect to SOCKS5 proxies. It
// uses a base dialer under the hood.
func NewSocks5(opt Opts, proxyAuth ProxyAuth) (Dialer, error) {
	baseDialer, _ := NewBase(opt).(*base)

	var auth *proxy.Auth

	if proxyAuth.HasCredentials() {
		auth = &proxy.Auth{
			User:     proxyAuth.Username,
			Password: proxyAuth.Password,
		}
	}

	proxyInstance, err := proxy.SOCKS5("tcp", proxyAuth.Address, auth, StdDialerWrapper{baseDialer})
	if err != nil {
		return nil, fmt.Errorf("cannot build socks5 proxy: %w", err)
	}

	return &socksProxy{
		baseDialer: baseDialer,
		proxy:      proxyInstance.(proxy.ContextDialer),
	}, nil
}
