package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/ca"
	"github.com/9seconds/httransform/v2/conns"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/headers"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/valyala/fasthttp"
)

type Server struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	serverPool    sync.Pool
	eventStream   events.Stream
	layers        []layers.Layer
	authenticator auth.Interface
	executor      executor.Executor
	ca            *ca.CA
	server        *fasthttp.Server
}

func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

func (s *Server) Close() error {
	s.ctxCancel()

	return s.server.Shutdown()
}

func (s *Server) entrypoint(ctx *fasthttp.RequestCtx) {
	user, err := s.authenticator.Authenticate(ctx)
	if err != nil {
		ctx.Error(fmt.Sprintf("authenication is failed: %v", err), fasthttp.StatusProxyAuthRequired)
		ctx.Response.Header.Add("Proxy-Authenticate", "Basic")
		s.eventStream.Send(ctx, events.EventTypeFailedAuth, nil, "")

		return
	}

	var requestType events.RequestType

	if ctx.IsConnect() {
		requestType |= events.RequestTypeTunneled

		address, err := s.extractAddress(string(ctx.RequestURI()), true)
		if err != nil {
			ctx.Error(fmt.Sprintf("cannot extract a host for tunneled connection: %s", err.Error()), fasthttp.StatusBadRequest)

			return
		}

		ctx.Hijack(s.upgradeToTLS(requestType, user, address))
		ctx.Success("", nil)

		return
	}

	address, err := s.extractAddress(string(ctx.Host()), false)
	if err != nil {
		ctx.Error(fmt.Sprintf("cannot extract a host for tunneled connection: %s", err.Error()), fasthttp.StatusBadRequest)

		return
	}

	ownCtx := layers.AcquireContext()
	defer layers.ReleaseContext(ownCtx)

	ownCtx.Init(ctx, address, s.eventStream, user, requestType)
	s.completeRequestType(ownCtx)
	s.main(ownCtx)
}

func (s *Server) upgradeToTLS(requestType events.RequestType, user, address string) fasthttp.HijackHandler {
	host, _, _ := net.SplitHostPort(address)

	return conns.FixHijackHandler(func(conn net.Conn) bool {
		conf, err := s.ca.Get(host)
		if err != nil {
			return true
		}

		uConn := conns.NewUnreadConn(conn)
		tlsErr := tls.RecordHeaderError{}
		tlsConn := tls.Server(uConn, conf)
		err = tlsConn.Handshake()

		switch {
		case errors.As(err, &tlsErr) && tlsErr.Conn != nil:
			uConn.Unread()
		case err != nil:
			return true
		default:
			requestType |= events.RequestTypeTLS
			uConn.Seal()
		}

		srv := s.serverPool.Get().(*fasthttp.Server)
		defer s.serverPool.Put(srv)

		needToClose := true

		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			ownCtx := layers.AcquireContext()
			defer layers.ReleaseContext(ownCtx)

			ownCtx.Init(ctx, address, s.eventStream, user, requestType)
			s.completeRequestType(ownCtx)
			s.main(ownCtx)

			needToClose = !ownCtx.Hijacked()
		}

		srv.ServeConn(tlsConn) // nolint: errcheck

		return needToClose
	})
}

func (s *Server) main(ctx *layers.Context) {
	requestMeta := &events.RequestMeta{
		RequestID:   ctx.RequestID,
		RequestType: ctx.RequestType,
		Method:      string(bytes.ToUpper(ctx.Request().Header.Method())),
		User:        ctx.User,
		Addr:        ctx.RemoteAddr(),
	}

	ctx.Request().URI().CopyTo(&requestMeta.URI)
	s.eventStream.Send(s.ctx, events.EventTypeStartRequest, requestMeta, ctx.RequestID)

	defer func() {
		responseMeta := &events.ResponseMeta{
			RequestID:  ctx.RequestID,
			StatusCode: ctx.Response().StatusCode(),
		}

		s.eventStream.Send(s.ctx, events.EventTypeFinishRequest, responseMeta, ctx.RequestID)
	}()

	currentLayer := 0

	var err error

	for ; err == nil && currentLayer < len(s.layers); currentLayer++ {
		err = s.layers[currentLayer].OnRequest(ctx)
	}

	if err == nil {
		err = s.executor(ctx)
	}

	for currentLayer--; currentLayer >= 0; currentLayer-- {
		err = s.layers[currentLayer].OnResponse(ctx, err)
	}

	if err != nil {
		errorMeta := &events.ErrorMeta{
			RequestID: ctx.RequestID,
			Err:       err,
		}

		s.eventStream.Send(ctx, events.EventTypeFailedRequest, errorMeta, ctx.RequestID)
		ctx.Error(err)
	}
}

func (s *Server) extractAddress(hostport string, isTLS bool) (string, error) {
	_, _, err := net.SplitHostPort(hostport)

	var addrError *net.AddrError

	switch {
	case errors.As(err, &addrError) && strings.Contains(addrError.Err, "missing port"):
		if isTLS {
			return net.JoinHostPort(hostport, "443"), nil
		}

		return net.JoinHostPort(hostport, "80"), nil
	case err != nil:
		return "", fmt.Errorf("incorrect host %s: %w", hostport, err)
	}

	return hostport, nil
}

func (s *Server) completeRequestType(ctx *layers.Context) {
	ctx.Request().Header.VisitAll(func(key, value []byte) {
		if bytes.EqualFold(key, []byte("Connection")) {
			values := headers.Values(string(value))
			ctx.RequestType &^= events.RequestTypeUpgraded

			for i := range values {
				if strings.EqualFold(values[i], "Upgrade") {
					ctx.RequestType |= events.RequestTypeUpgraded
				}
			}
		}
	})
}

func NewServer(ctx context.Context, opts ServerOpts) (*Server, error) { // nolint: funlen
	ctx, cancel := context.WithCancel(ctx)
	oopts := &opts
	eventStream := events.NewStream(ctx, oopts.GetEventProcessorFactory())
	authenticator := oopts.GetAuthenticator()

	certAuth, err := ca.NewCA(ctx,
		eventStream,
		oopts.GetTLSCertCA(),
		oopts.GetTLSPrivateKey())
	if err != nil {
		cancel()

		return nil, fmt.Errorf("cannot make certificate authority: %w", err)
	}

	exec := oopts.GetExecutor()
	if exec == nil {
		dialer := dialers.NewBase(dialers.Opts{
			TLSSkipVerify: oopts.GetTLSSkipVerify(),
		})
		exec = executor.MakeDefaultExecutor(dialer)
	}

	srv := &Server{
		ctx:           ctx,
		ctxCancel:     cancel,
		eventStream:   eventStream,
		ca:            certAuth,
		layers:        oopts.GetLayers(),
		authenticator: authenticator,
		executor:      exec,
		serverPool: sync.Pool{
			New: func() interface{} {
				return &fasthttp.Server{
					Concurrency:                   oopts.GetConcurrency(),
					ReadBufferSize:                oopts.GetReadBufferSize(),
					WriteBufferSize:               oopts.GetWriteBufferSize(),
					ReadTimeout:                   oopts.GetReadTimeout(),
					WriteTimeout:                  oopts.GetWriteTimeout(),
					TCPKeepalivePeriod:            oopts.GetTCPKeepAlivePeriod(),
					MaxRequestBodySize:            oopts.GetMaxRequestBodySize(),
					DisableKeepalive:              false,
					TCPKeepalive:                  true,
					DisableHeaderNamesNormalizing: true,
					NoDefaultServerHeader:         true,
					NoDefaultContentType:          true,
					NoDefaultDate:                 true,
					DisablePreParseMultipartForm:  true,
					KeepHijackedConns:             true,
					ErrorHandler: func(cctx *fasthttp.RequestCtx, err error) {
						meta := &events.CommonErrorMeta{
							Method: string(bytes.ToUpper(cctx.Method())),
							Addr:   cctx.RemoteAddr(),
							Err:    err,
						}

						cctx.URI().CopyTo(&meta.URI)
						eventStream.Send(ctx, events.EventTypeCommonError, meta, "")
					},
				}
			},
		},
	}
	srv.server = srv.serverPool.Get().(*fasthttp.Server)
	srv.server.Handler = srv.entrypoint

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return srv, nil
}
