package httransform

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/ca"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/valyala/fasthttp"
)

type Server struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	serverPool    sync.Pool
	channelEvents chan<- events.Event
	layers        []layers.Layer
	auth          []byte
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
		s.sendEvent(events.EventTypeFailedAuth, nil)

		return
	}

	if ctx.IsConnect() {
		host, _, err := net.SplitHostPort(string(ctx.RequestURI()))
		if err != nil {
			ctx.Error(fmt.Sprintf("cannot extract a host for tunneled connection: %s", err.Error()), fasthttp.StatusBadRequest)

			return
		}

		ctx.Hijack(s.hijackConnect(host, ctx))
		ctx.Success("", nil)

		return
	}

	ownCtx := layers.AcquireContext()
	defer layers.ReleaseContext(ownCtx)
	defer ownCtx.Cancel()

	ownCtx.Init(ctx, false)
	s.sendEvent(events.EventTypePlainRequest, ownCtx.RequestID())
	s.main(ownCtx)
}

func (s *Server) hijackConnect(host string, ctx *fasthttp.RequestCtx) fasthttp.HijackHandler {
	return func(conn net.Conn) {
		defer conn.Close()

		conf, err := s.ca.Get(host)
		if err != nil {
			return
		}

		tlsConn := tls.Server(conn, conf)
		defer tlsConn.Close()

		if err := tlsConn.Handshake(); err != nil {
			return
		}

		srv := s.serverPool.Get().(*fasthttp.Server)
		defer s.serverPool.Put(srv)

		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			ownCtx := layers.AcquireContext()
			defer layers.ReleaseContext(ownCtx)
			defer ownCtx.Cancel()

			ownCtx.Init(ctx, true)
			s.sendEvent(events.EventTypeTunneledRequest, ownCtx.RequestID())
			s.main(ownCtx)
		}

		srv.ServeConn(tlsConn)
	}
}

func (s *Server) main(ctx *layers.Context) {
	defer s.sendEvent(events.EventTypeRequestDone, ctx.RequestID())

	ctx.Respond("YES", 200)
}

func (s *Server) sendEvent(eventType events.EventType, value interface{}) {
	select {
	case <-s.ctx.Done():
	case s.channelEvents <- events.New(eventType, value):
	}
}

func NewServer(ctx context.Context, opts ServerOpts) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	oopts := &opts
	channelEvents := events.NewEventChannel(ctx, oopts.GetEventProcessor())
	proxyAuth := oopts.GetAuth()
	certAuth, err := ca.NewCA(ctx, channelEvents,
		oopts.GetTLSCertCA(),
		oopts.GetTLSPrivateKey(),
		oopts.GetTLSOrgName(),
		oopts.GetTLSCacheSize())

	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot make certificate authority: %w", err)
	}

	srv := &Server{
		ctx:           ctx,
		ctxCancel:     cancel,
		channelEvents: channelEvents,
		ca:            certAuth,
		layers:        oopts.GetLayers(),
		auth:          proxyAuth,
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
