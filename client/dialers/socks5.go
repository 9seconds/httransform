package dialers

import (
	"context"
	"net"
	"time"

	"github.com/PumpkinSeed/errors"
	"github.com/libp2p/go-reuseport"
	"golang.org/x/net/proxy"
)

type socksProxyDialer interface {
	ProxyDial(context.Context, string, string) (net.Conn, error)
}

type socksDialer struct {
	dialer  socksProxyDialer
	timeout time.Duration
}

func (s *socksDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	conn, err := s.dialer.ProxyDial(ctx, "tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, ErrDialer)
	}

	return conn, nil
}

func (s *socksDialer) GetTimeout() time.Duration {
	return s.timeout
}

func NewSocks5(address string, timeout time.Duration, user, password string) (Dialer, error) {
	var auth *proxy.Auth
	if user != "" || password != "" {
		auth = &proxy.Auth{
			User:     user,
			Password: password,
		}
	}

	socks, err := proxy.SOCKS5("tcp", address, auth, &net.Dialer{
		Timeout: timeout,
		Control: reuseport.Control,
	})
	if err != nil {
		return nil, err
	}

	return &socksDialer{
		dialer:  socks.(socksProxyDialer),
		timeout: timeout,
	}, nil
}
