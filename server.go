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
	"github.com/9seconds/httransform/v2/layers"
	"github.com/valyala/fasthttp"
)

type Server struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	serverPool    sync.Pool
	channelEvents events.EventChannel
	layers        []layers.Layer
	auth          []byte
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
	if err := auth.AuthenticateRequestHeaders(&ctx.Request.Header, s.auth); err != nil {
		ctx.Error(fmt.Sprintf("Cannot authenticate: %s", err.Error()), fasthttp.StatusProxyAuthRequired)
		s.channelEvents.Send(ctx, events.EventTypeFailedAuth, nil, "")

		return
	}

	if ctx.IsConnect() {
		address, err := s.extractAddress(string(ctx.RequestURI()), true)
		if err != nil {
			ctx.Error(fmt.Sprintf("cannot extract a host for tunneled connection: %s", err.Error()), fasthttp.StatusBadRequest)

			return
		}

		ctx.Hijack(s.upgradeToTLS(address))
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

	ownCtx.Init(ctx, address, s.channelEvents, false)
	s.main(ownCtx)
}

func (s *Server) upgradeToTLS(address string) fasthttp.HijackHandler {
	host, _, _ := net.SplitHostPort(address)

	return conns.FixedHijackHandler(func(conn net.Conn) bool {
		conf, err := s.ca.Get(host)
		if err != nil {
			return true
		}

		tlsConn := tls.Server(conn, conf)
		if err := tlsConn.Handshake(); err != nil {
			return true
		}

		srv := s.serverPool.Get().(*fasthttp.Server)
		defer s.serverPool.Put(srv)

		needToClose := true

		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			ownCtx := layers.AcquireContext()
			defer layers.ReleaseContext(ownCtx)

			ownCtx.Init(ctx, address, s.channelEvents, true)
			s.main(ownCtx)

			needToClose = !ownCtx.Hijacked()
		}

		srv.ServeConn(tlsConn) // nolint: errcheck

		return needToClose
	})
}

func (s *Server) main(ctx *layers.Context) {
	requestMeta := &events.RequestMeta{
		RequestID: ctx.RequestID,
		Method:    string(bytes.ToLower(ctx.Request().Header.Method())),
	}

	ctx.Request().URI().CopyTo(&requestMeta.URI)
	s.channelEvents.Send(s.ctx, events.EventTypeStartRequest, requestMeta, ctx.RequestID)

	defer func() {
		responseMeta := &events.ResponseMeta{
			RequestID:  ctx.RequestID,
			StatusCode: ctx.Response().StatusCode(),
		}

		s.channelEvents.Send(s.ctx, events.EventTypeFinishRequest, responseMeta, ctx.RequestID)
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
			Error:     err,
		}

		s.channelEvents.Send(ctx, events.EventTypeFailedRequest, errorMeta, ctx.RequestID)
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

func NewServer(ctx context.Context, opts ServerOpts) (*Server, error) { // nolint: funlen
	ctx, cancel := context.WithCancel(ctx)
	oopts := &opts
	channelEvents := events.NewEventChannel(ctx, oopts.GetEventProcessor())
	proxyAuth := oopts.GetAuth()

	certAuth, err := ca.NewCA(ctx, channelEvents,
		oopts.GetTLSCertCA(),
		oopts.GetTLSPrivateKey(),
		oopts.GetTLSCacheSize())
	if err != nil {
		cancel()

		return nil, fmt.Errorf("cannot make certificate authority: %w", err)
	}

	exec := oopts.GetExecutor()
	if exec == nil {
		dialer, _ := dialers.NewBase(dialers.Opts{
			Context:       ctx,
			TLSSkipVerify: oopts.GetTLSSkipVerify(),
		})
		exec = executor.MakeDefaultExecutor(dialer)
	}

	srv := &Server{
		ctx:           ctx,
		ctxCancel:     cancel,
		channelEvents: channelEvents,
		ca:            certAuth,
		layers:        oopts.GetLayers(),
		auth:          proxyAuth,
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
				}
			},
		},
	}
	srv.server = srv.serverPool.Get().(*fasthttp.Server)
	srv.server.Handler = srv.entrypoint
	srv.server.ContinueHandler = func(headers *fasthttp.RequestHeader) bool {
		return auth.AuthenticateRequestHeaders(headers, proxyAuth) == nil
	}

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return srv, nil
}
