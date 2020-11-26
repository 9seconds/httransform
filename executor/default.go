package executor

import (
	"bytes"
	"context"
	"fmt"
	"net"

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

		request := ctx.Request()
		response := ctx.Response()
		ownCtx, cancel := context.WithCancel(ctx)

		dialer.PatchHTTPRequest(request)

		go func() {
			<-ownCtx.Done()
			conn.Close()
		}()

		return http.Execute(ownCtx, conn, request, response, func() { cancel() })
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
		Parent:  conn,
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
