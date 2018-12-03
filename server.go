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
	logger     Logger
}

func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

func (s *Server) Shutdown() error {
	s.certs.Close()
	return s.server.Shutdown()
}

func (s *Server) mainHandler(ctx *fasthttp.RequestCtx) {
	var user []byte
	var password []byte
	var err error

	method := string(ctx.Method())
	uri := string(ctx.RequestURI())
	reqID := ctx.ID()

	proxyAuthHeaderValue := ctx.Request.Header.PeekBytes([]byte("Proxy-Authorization"))
	if len(proxyAuthHeaderValue) > 0 {
		user, password, err = ExtractAuthentication(proxyAuthHeaderValue)
		if err != nil {
			s.logger.Debug("[%s] (%d) %s %s: incorrect proxy-authorization header: %s",
				ctx.RemoteAddr(), reqID, method, uri, err)
			user = nil
			password = nil
		}
	}

	if ctx.IsConnect() {
		host, _, err := net.SplitHostPort(uri)
		if err != nil {
			s.logger.Debug("[%s] (%d) %s %s: cannot extract host for CONNECT: %s",
				ctx.RemoteAddr(), reqID, method, uri, err)
			MakeSimpleResponse(&ctx.Response, fmt.Sprintf("Cannot extract host for request URI %s", uri), fasthttp.StatusBadRequest)
			return
		}
		ctx.Hijack(s.makeHijackHandler(host, reqID, user, password))
		ctx.Success("", nil)
		return
	}

	s.handleRequest(ctx, false, user, password)
}

func (s *Server) makeHijackHandler(host string, reqID uint64, user, password []byte) fasthttp.HijackHandler {
	return func(conn net.Conn) {
		defer conn.Close()

		conf, err := s.certs.Get(host)
		if err != nil {
			s.logger.Warn("[%s] (%d): cennot generate certificate for the host %s: %s",
				conn.RemoteAddr(), reqID, host, err)
			return
		}
		defer conf.Release()

		tlsConn := tls.Server(conn, conf.Get())
		defer tlsConn.Close()

		if err = tlsConn.Handshake(); err != nil {
			s.logger.Warn("[%s] (%d): cennot finish TLS handshake: %s",
				conn.RemoteAddr(), reqID, err)
			return
		}

		srv := s.serverPool.Get().(*fasthttp.Server)
		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			s.handleRequest(ctx, true, user, password)
		}
		defer s.serverPool.Put(srv)

		srv.ServeConn(tlsConn) // nolint: errcheck
	}
}

func (s *Server) handleRequest(ctx *fasthttp.RequestCtx, isConnect bool, user, password []byte) {
	reqID := ctx.ID()
	method := string(ctx.Method())
	uri := string(ctx.RequestURI())

	requestHeaders := getHeaderSet()
	responseHeaders := getHeaderSet()
	defer func() {
		releaseHeaderSet(requestHeaders)
		releaseHeaderSet(responseHeaders)
	}()

	if err := parseHeaders(requestHeaders, ctx.Request.Header.Header()); err != nil {
		s.logger.Debug("[%s] (%d) %s %s: malformed request headers: %s",
			ctx.RemoteAddr(), reqID, method, uri, err)
		MakeSimpleResponse(&ctx.Response, "Malformed request headers", fasthttp.StatusBadRequest)
		return
	}

	state := getLayerState()
	defer releaseLayerState(state)
	initLayerState(state, ctx, requestHeaders, responseHeaders, isConnect, user, password)

	requestMethod := ctx.Request.Header.Method()
	requestURI := ctx.Request.Header.RequestURI()
	currentLayer := 0
	var err error

	for ; currentLayer < len(s.layers); currentLayer++ {
		if err = s.layers[currentLayer].OnRequest(state); err != nil {
			MakeSimpleResponse(&ctx.Response, "Internal Server Error", fasthttp.StatusInternalServerError)
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
	if err2 := parseHeaders(responseHeaders, state.Response.Header.Header()); err2 != nil {
		s.logger.Debug("[%s] (%d) %s %s: malformed response headers: %s",
			ctx.RemoteAddr(), reqID, method, uri, err)
		MakeSimpleResponse(&ctx.Response, "Malformed response headers", fasthttp.StatusBadRequest)
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

func NewServer(opts ServerOpts, lrs []Layer, executor Executor, logger Logger) (*Server, error) {
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
		logger:   logger,
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
