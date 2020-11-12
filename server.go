package httransform

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/9seconds/httransform/v2/ca"
	"github.com/9seconds/httransform/v2/events"
	"github.com/valyala/fasthttp"
)

type Server struct {
	ctx         context.Context
	ctxCancel   context.CancelFunc
	serverPool  sync.Pool
	eventStream *events.EventStream
	ca          *ca.CA
	server      *fasthttp.Server
}

func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

func (s *Server) Close() error {
	s.ctxCancel()

	return s.server.Shutdown()
}

func (s *Server) main(ctx *fasthttp.RequestCtx) {

}

func NewServer(ctx context.Context, opts ServerOpts) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	oopts := &opts
	eventStream := events.NewEventStream(ctx, oopts.GetEventProcessor())
	certAuth, err := ca.NewCA(ctx, eventStream,
		oopts.GetTLSCertCA(),
		oopts.GetTLSPrivateKey(),
		oopts.GetTLSOrgName(),
		oopts.GetTLSCacheSize())

	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot make certificate authority: %w", err)
	}

	srv := &Server{
		ctx:         ctx,
		ctxCancel:   cancel,
		eventStream: eventStream,
		ca:          certAuth,
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
	srv.server.Handler = srv.main

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return srv, nil
}
