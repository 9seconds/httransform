package dialer

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/proxy"
)

type socksProxy struct {
	proxy proxy.ContextDialer
}

func (s *socksProxy) Dial(network, address string) (net.Conn, error) {
	return s.proxy.DialContext(context.Background(), network, address)
}

func (s *socksProxy) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return s.proxy.DialContext(ctx, network, address)
}

func NewSocks(ctx context.Context, dialTimeout time.Duration, address, username, password string) (Dialer, error) {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return nil, fmt.Errorf("incorrect address: %w", err)
	}

	var auth *proxy.Auth

	if username != "" || password != "" {
		auth = &proxy.Auth{
			User:     username,
			Password: password,
		}
	}

	proxyInstance, err := proxy.SOCKS5("tcp", address, auth, NewBase(ctx, dialTimeout))
	if err != nil {
		return nil, fmt.Errorf("cannot build socks5 proxy: %w", err)
	}

	return &socksProxy{
		proxy: proxyInstance.(proxy.ContextDialer),
	}, nil
}
