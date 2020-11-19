package dialers

import (
	"context"
	"fmt"
	"net"

	"github.com/valyala/fasthttp"
)

type Dialer interface {
	Dial(context.Context, string, string) (net.Conn, error)
	UpgradeToTLS(context.Context, net.Conn, string) (net.Conn, error)
	PatchHTTPRequest(*fasthttp.Request)
}

func Dial(ctx context.Context, dialer Dialer, address string, upgradeToTLS bool) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("incorrect address format: %w", err)
	}

	conn, err := dialer.Dial(ctx, host, port)
	if err != nil {
		return nil, fmt.Errorf("cannot establish tcp connection to %s: %w", address, err)
	}

	if !upgradeToTLS {
		return conn, nil
	}

	tlsConn, err := dialer.UpgradeToTLS(ctx, conn, host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("cannot upgrade connection to %s to tls: %w", address, err)
	}

	return tlsConn, nil
}
