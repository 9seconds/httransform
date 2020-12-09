package executor

import (
	"bytes"
	"io"
	"net"
	"strings"

	"github.com/9seconds/httransform/v2/conns"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/errors"
	"github.com/9seconds/httransform/v2/http"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/9seconds/httransform/v2/upgrades"
)

func MakeDefaultExecutor(dialer dialers.Dialer) Executor {
	return func(ctx *layers.Context) error {
		conn, err := defaultExecutorDial(ctx, dialer)
		if err != nil {
			return errors.Annotate(err, "cannot dial to the netloc", "", 0)
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
		return nil, errors.Annotate(err, "incorrect address format", "", 0)
	}

	conn, err := dialer.Dial(ctx, host, port)
	if err != nil {
		return nil, errors.Annotate(err, "cannot establish tcp connection", "", 0)
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

		return nil, errors.Annotate(err, "cannot upgrade connection to tls", "", 0)
	}

	return tlsConn, nil
}

func defaultExecutorConnectionUpgrade(ctx *layers.Context, conn net.Conn) error {
	if err := defaultExecutorHTTPRequest(ctx, conn); err != nil {
		return errors.Annotate(err, "cannot upgrade http connection", "", 0)
	}

	ctx.Hijack(conn, func(clientConn, netlocConn net.Conn) {
		upgrader := upgrades.AcquireTCP(upgrades.NoopTCPReactor{})
		defer upgrades.ReleaseTCP(upgrader)

		upgrader.Manage(ctx, clientConn, netlocConn)
	})

	return nil
}

func defaultExecutorHTTPRequest(ctx *layers.Context, conn io.ReadWriteCloser) error {
	if err := http.Execute(ctx, conn, ctx.Request(), ctx.Response()); err != nil {
		conn.Close()

		return errors.Annotate(err, "cannot send http request", "", 0)
	}

	return nil
}
