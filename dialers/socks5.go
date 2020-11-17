package dialers

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/net/proxy"
)

type socksProxy struct {
	baseDialer *base
	proxy      proxy.ContextDialer
}

func (s *socksProxy) Dial(ctx context.Context, host, port string) (net.Conn, error) {
	return s.proxy.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
}

func (s *socksProxy) UpgradeToTLS(ctx context.Context, conn net.Conn, host string) (net.Conn, error) {
	return s.baseDialer.UpgradeToTLS(ctx, conn, host)
}

func NewSocks(opt Opts, proxyAuth ProxyAuth) (Dialer, error) {
	baseDialer, err := NewBase(opt)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize a base dialer: %w", err)
	}

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
		baseDialer: baseDialer.(*base),
		proxy:      proxyInstance.(proxy.ContextDialer),
	}, nil
}
