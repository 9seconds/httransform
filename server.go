package httransform

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/ca"
)

type fasthttpHeader interface {
	Reset()
	DisableNormalizing()
	SetBytesKV([]byte, []byte)
}

type Server struct {
	serverPool sync.Pool
	server     *fasthttp.Server
	certs      ca.CA
	layers     []Layer
	executor   Executor
}

func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown()
}

func (s *Server) mainHandler(ctx *fasthttp.RequestCtx) {
	var user []byte
	var password []byte
	var err error

	proxyAuthHeaderValue := ctx.Request.Header.PeekBytes([]byte("Proxy-Authorization"))
	if len(proxyAuthHeaderValue) > 0 {
		user, password, err = ExtractAuthentication(proxyAuthHeaderValue)
		if err != nil {
			user = nil
			password = nil
		}
	}

	if ctx.IsConnect() {
		uri := string(ctx.RequestURI())
		host, err := ExtractHost(uri)
		if err != nil {
			MakeBadResponse(&ctx.Response, fmt.Sprintf("Cannot extract host for request URI %s", uri), fasthttp.StatusBadRequest)
			return
		}
		ctx.Hijack(s.makeHijackHandler(host, user, password))
		ctx.Success("", nil)
		return
	}

	ctx.SetUserValue("proxy_auth_user", user)
	ctx.SetUserValue("proxy_auth_password", password)

	s.handleRequest(ctx, false)
}

func (s *Server) makeHijackHandler(host string, user, password []byte) fasthttp.HijackHandler {
	return func(conn net.Conn) {
		defer conn.Close()

		conf, err := s.certs.Get(host)
		if err != nil {
			return
		}
		defer conf.Release()

		tlsConn := tls.Server(conn, conf.Get())
		defer tlsConn.Close()

		if err = tlsConn.Handshake(); err != nil {
			return
		}

		srv := s.serverPool.Get().(*fasthttp.Server)
		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.SetUserValue("proxy_auth_user", user)
			ctx.SetUserValue("proxy_auth_password", password)
			s.handleRequest(ctx, true)
		}
		defer s.serverPool.Put(srv)

		srv.ServeConn(tlsConn)
	}
}

func (s *Server) handleRequest(ctx *fasthttp.RequestCtx, isConnect bool) {
	requestHeaders := getHeaderSet()
	responseHeaders := getHeaderSet()
	defer func() {
		requestHeaders.Clear()
		responseHeaders.Clear()
		releaseHeaderSet(requestHeaders)
		releaseHeaderSet(responseHeaders)
	}()

	if err := parseHeaders(requestHeaders, ctx.Request.Header.Header()); err != nil {
		MakeBadResponse(&ctx.Response, "Malformed request headers", fasthttp.StatusBadRequest)
		return
	}

	state := getLayerState()
	defer releaseLayerState(state)
	initLayerState(state, ctx, requestHeaders, responseHeaders, isConnect)

	requestMethod := ctx.Request.Header.Method()
	requestURI := ctx.Request.Header.RequestURI()
	currentLayer := 0
	var err error

	for ; currentLayer < len(s.layers); currentLayer++ {
		if err = s.layers[currentLayer].OnRequest(state); err != nil {
			MakeBadResponse(&ctx.Response, "Internal Server Error", fasthttp.StatusInternalServerError)
			break
		}
	}

	if currentLayer == len(s.layers) {
		currentLayer--
		s.resetHeaders(&ctx.Request.Header, requestHeaders)
		ctx.Request.Header.SetMethodBytes(requestMethod)
		ctx.Request.Header.SetRequestURIBytes(requestURI)

		s.executor(state)
	}
	if err := parseHeaders(responseHeaders, state.Response.Header.Header()); err != nil {
		MakeBadResponse(&ctx.Response, "Malformed response headers", fasthttp.StatusBadRequest)
		return
	}

	for ; currentLayer >= 0; currentLayer-- {
		s.layers[currentLayer].OnResponse(state, err)
	}
	responseCode := ctx.Response.Header.StatusCode()
	s.resetHeaders(&ctx.Response.Header, responseHeaders)
	ctx.Response.SetStatusCode(responseCode)
}

func (s *Server) resetHeaders(headers fasthttpHeader, set *HeaderSet) {
	headers.Reset()
	headers.DisableNormalizing()
	for _, v := range set.Items() {
		headers.SetBytesKV(v.Key, v.Value)
	}
}

func NewServer(opts ServerOpts, lrs []Layer, executor Executor) (*Server, error) {
	certs, err := ca.NewCA(opts.GetCertCA(),
		opts.GetCertKey(),
		opts.GetTLSCertCacheSize(),
		opts.GetTLSCertCachePrune(),
		opts.GetOrganizationName())
	if err != nil {
		return nil, errors.Annotate(err, "Cannot create CA")
	}

	srv := &Server{
		certs:    certs,
		layers:   lrs,
		executor: executor,
	}
	srv.serverPool = sync.Pool{
		New: func() interface{} {
			return &fasthttp.Server{
				Concurrency:                   opts.GetConcurrency(),
				ReadBufferSize:                opts.GetReadBufferSize(),
				WriteBufferSize:               opts.GetWriteBufferSize(),
				ReadTimeout:                   opts.GetReadTimeout(),
				WriteTimeout:                  opts.GetWriteTimeout(),
				DisableKeepalive:              false,
				TCPKeepalive:                  true,
				DisableHeaderNamesNormalizing: true,
				NoDefaultServerHeader:         true,
				NoDefaultContentType:          true,
			}
		},
	}
	srv.server = srv.serverPool.Get().(*fasthttp.Server)
	srv.server.Handler = srv.mainHandler

	return srv, nil
}
