package httransform

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/PumpkinSeed/errors"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/ca"
	"github.com/9seconds/httransform/v2/layers"
)

type Server struct {
	ctx        context.Context
	cancel     context.CancelFunc
	auth       auth.Auth
	serverPool sync.Pool
	server     *fasthttp.Server
	ca         *ca.CA
	layers     []layers.Layer
}

func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

func (s *Server) Shutdown() error {
	defer s.ca.Close()

	s.cancel()

	return s.server.Shutdown()
}

func (s *Server) handleRequest(ctx *layers.LayerContext) {
	var err error

	currentLayer := 0
	for ; currentLayer <= len(s.layers); currentLayer++ {
		err = s.layers[currentLayer].OnRequest(ctx)
		if err != nil {
			ctx.SetSimpleResponse(fasthttp.StatusInternalServerError, "Internal server error")
			break
		}
	}

	if currentLayer == len(s.layers) {
		currentLayer--

		ctx.SyncToRequest()
	}

	for ; currentLayer >= 0; currentLayer-- {
		s.layers[currentLayer].OnResponse(ctx, err)
	}

	ctx.SyncToResponse()
}

func (s *Server) main(ctx *fasthttp.RequestCtx) {
	if ctx.IsConnect() {
		s.processConnect(ctx)
		return
	}

	layerCtx := layers.AcquireLayerContext()
	defer layers.ReleaseLayerContext(layerCtx)

	if err := layers.InitLayerContext(ctx, layerCtx); err != nil {
		ctx.Error(fmt.Sprintf("Cannot process this request: %s", err),
			fasthttp.StatusUnprocessableEntity)
		return
	}

	s.handleRequest(layerCtx)
}

func (s *Server) processConnect(ctx *fasthttp.RequestCtx) {
	connectHostPort := bytes.TrimPrefix(ctx.Request.URI().Path(), []byte("/"))
	host, portStr, err := net.SplitHostPort(string(connectHostPort))

	if err != nil {
		ctx.Error(fmt.Sprintf("Cannot process this request: %s", err),
			fasthttp.StatusUnprocessableEntity)
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		ctx.Error(fmt.Sprintf("Cannot process this request: %s", err),
			fasthttp.StatusUnprocessableEntity)
		return
	}

	tlsConfig, err := s.ca.Get(host)
	if err != nil {
		ctx.Error(fmt.Sprintf("Cannot generate TLS certificate: %s", err),
			fasthttp.StatusInternalServerError)
		return
	}

	layerCtx := layers.AcquireLayerContext()
	defer layers.ReleaseLayerContext(layerCtx)

	if err := layers.InitLayerContext(ctx, layerCtx); err != nil {
		ctx.Error(fmt.Sprintf("Cannot process this request: %s", err),
			fasthttp.StatusUnprocessableEntity)
		return
	}

	if _, err := s.auth.Auth(layerCtx); err != nil {
		ctx.Error(fmt.Sprintf("Cannot process this request: %s", err),
			fasthttp.StatusProxyAuthRequired)
		return
	}

	ctx.Hijack(func(conn net.Conn) {
		defer conn.Close()

		tlsConn := tls.Server(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			return
		}
		defer tlsConn.Close()

		srv := s.serverPool.Get().(*fasthttp.Server)
		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			layerCtx := layers.AcquireLayerContext()
			defer layers.ReleaseLayerContext(layerCtx)

			if err := layers.InitLayerContext(ctx, layerCtx); err != nil {
				ctx.Error(fmt.Sprintf("Cannot process this request: %s", err),
					fasthttp.StatusUnprocessableEntity)
				return
			}

			layerCtx.ConnectHost = host
			layerCtx.ConnectPort = port

			s.handleRequest(layerCtx)
		}
		defer s.serverPool.Put(srv)

		srv.ServeConn(tlsConn)
	})

	ctx.Success("text/plain", nil)
}

func NewServer(ctx context.Context, opts ServerOpts) (*Server, error) {
	certs, err := ca.NewCA(ctx,
		opts.GetCertCA(),
		opts.GetCertKey(),
		opts.GetTLSCertCacheSize(),
		opts.GetOrganizationName())

	if err != nil {
		return nil, errors.Wrap(errors.New("cannot create CA"), err)
	}

	ctx, cancel := context.WithCancel(ctx)
	popts := &opts

	srv := &Server{
		ctx:    ctx,
		cancel: cancel,
		ca:     certs,
		layers: opts.GetLayers(),
		auth:   opts.GetAuth(),
		serverPool: sync.Pool{
			New: func() interface{} {
				return &fasthttp.Server{
					Concurrency:                   popts.GetConcurrency(),
					ReadBufferSize:                popts.GetReadBufferSize(),
					WriteBufferSize:               popts.GetWriteBufferSize(),
					ReadTimeout:                   popts.GetReadTimeout(),
					WriteTimeout:                  popts.GetWriteTimeout(),
					DisableKeepalive:              false,
					TCPKeepalive:                  true,
					DisableHeaderNamesNormalizing: true,
					NoDefaultServerHeader:         true,
					NoDefaultContentType:          true,
				}
			},
		},
	}
	srv.server = srv.serverPool.Get().(*fasthttp.Server)
	srv.server.Handler = srv.main

	return srv, nil
}
