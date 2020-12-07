package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/9seconds/httransform/v2/conns"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/http"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/9seconds/httransform/v2/upgrades"
)

func MakeDefaultExecutor(dialer dialers.Dialer) Executor {
	return func(ctx *layers.Context) error {
		conn, err := defaultExecutorDial(ctx, dialer)
		if err != nil {
			return fmt.Errorf("cannot dial to the netloc: %w", err)
		}

		dialer.PatchHTTPRequest(ctx.Request())

		for _, v := range ctx.RequestHeaders.GetLast("Connection").Values() {
			if strings.EqualFold(v, "Upgrade") {
				return defaultExecutorConnectionUpgrade(ctx, conn)
			}
		}

		return defaultExecutorHTTPRequest(ctx, conn)
	}
}

func defaultExecutorDial(ctx *layers.Context, dialer dialers.Dialer) (net.Conn, error) {
	host, port, err := net.SplitHostPort(ctx.ConnectTo)
	if err != nil {
		return nil, fmt.Errorf("incorrect address format: %w", err)
	}

	conn, err := dialer.Dial(ctx, host, port)
	if err != nil {
		return nil, fmt.Errorf("cannot establish tcp connection: %w", err)
	}

	conn = &conns.TrafficConn{
		Conn:        conn,
		Context:     ctx,
		ID:          ctx.RequestID,
		EventStream: ctx.EventStream,
	}

	if bytes.EqualFold(ctx.Request().URI().Scheme(), []byte("http")) {
		return conn, nil
	}

	tlsConn, err := dialer.UpgradeToTLS(ctx, conn, host, port)
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf("cannot upgrade connection to tls: %w", err)
	}

	return tlsConn, nil
}

func defaultExecutorConnectionUpgrade(ctx *layers.Context, conn net.Conn) error {
	if err := http.Execute(ctx, conn, ctx.Request(), ctx.Response(), http.NoopResponseCallback); err != nil {
		return fmt.Errorf("cannot send http request: %w", err)
	}

	ctx.Hijack(conn, func(clientConn, netlocConn net.Conn) {
		upgrader := upgrades.AcquireTCP(upgrades.NoopTCPReactor{})
		defer upgrades.ReleaseTCP(upgrader)

		upgrader.Manage(ctx, clientConn, netlocConn)
	})

	return nil
}

func defaultExecutorHTTPRequest(ctx *layers.Context, conn io.ReadWriteCloser) error {
	ownCtx, cancel := context.WithCancel(ctx)

	go func() {
		<-ownCtx.Done()
		conn.Close()
	}()

	return http.Execute(ownCtx, conn, ctx.Request(), ctx.Response(), func() { cancel() })
}
