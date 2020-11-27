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
)

func MakeDefaultExecutor(dialer dialers.Dialer) Executor {
	return func(ctx *layers.Context) error {
		conn, err := defaultExecutorDial(ctx, dialer)
		if err != nil {
			return fmt.Errorf("cannot dial to the netloc: %w", err)
		}

		dialer.PatchHTTPRequest(ctx.Request())

		header := ctx.RequestHeaders.GetLast("Connection")
		if header != nil && strings.EqualFold(header.Value, "Upgrade") {
			return defaultExecutorConnectionUpgrade(ctx, conn)
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
		Conn:    conn,
		Context: ctx,
		ID:      ctx.RequestID,
		Events:  ctx.Events,
	}

	if bytes.EqualFold(ctx.Request().URI().Scheme(), []byte("http")) {
		return conn, nil
	}

	tlsConn, err := dialer.UpgradeToTLS(ctx, conn, host)
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
		go tcpPipe(clientConn, netlocConn)

		tcpPipe(netlocConn, clientConn)
	})

	return nil
}

func defaultExecutorHTTPRequest(ctx *layers.Context, conn net.Conn) error {
	ownCtx, cancel := context.WithCancel(ctx)

	go func() {
		<-ownCtx.Done()
		conn.Close()
	}()

	return http.Execute(ownCtx, conn, ctx.Request(), ctx.Response(), func() { cancel() })
}

func tcpPipe(src io.ReadCloser, dst io.WriteCloser) {
	defer func() {
		src.Close()
		dst.Close()
	}()

	io.Copy(dst, src)
}
