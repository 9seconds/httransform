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
	ContentLength() int
	SetContentLength(int)
}

// Server presents a HTTP proxy service.
type Server struct {
	serverPool sync.Pool
	tracerPool *TracerPool
	server     *fasthttp.Server
	certs      ca.CA
	layers     []Layer
	executor   Executor
	logger     Logger
	metrics    Metrics
}

// Serve starts to work using given listener.
func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	defer s.certs.Close()
	return s.server.Shutdown()
}

func (s *Server) mainHandler(ctx *fasthttp.RequestCtx) {
	var user []byte
	var password []byte
	var err error

	s.metrics.NewConnection()
	defer s.metrics.DropConnection()

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
		s.metrics.NewConnect()
		defer s.metrics.NewConnect()

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
	var err error

	currentLayer := 0
	reqID := ctx.ID()
	methodBytes := ctx.Method()
	method := string(methodBytes)
	uri := string(ctx.RequestURI())

	newMethodMetricsValue(s.metrics, methodBytes)
	requestHeaders := getHeaderSet()
	responseHeaders := getHeaderSet()
	defer func() {
		releaseHeaderSet(requestHeaders)
		releaseHeaderSet(responseHeaders)
		dropMethodMetricsValue(s.metrics, methodBytes)
	}()

	if err = ParseHeaders(requestHeaders, ctx.Request.Header.Header()); err != nil {
		s.logger.Debug("[%s] (%d) %s %s: malformed request headers: %s",
			ctx.RemoteAddr(), reqID, method, uri, err)
		MakeSimpleResponse(&ctx.Response, "Malformed request headers", fasthttp.StatusBadRequest)
		return
	}

	state := getLayerState()
	initLayerState(state, ctx, requestHeaders, responseHeaders, isConnect, user, password)
	layerTracer := s.tracerPool.acquire()
	defer func() {
		layerTracer.Dump(state, s.logger)
		s.tracerPool.release(layerTracer)
		releaseLayerState(state)
	}()

	for ; currentLayer < len(s.layers); currentLayer++ {
		layerTracer.StartOnRequest()
		err = s.layers[currentLayer].OnRequest(state)
		layerTracer.FinishOnRequest(err)
		if err != nil {
			MakeSimpleResponse(&ctx.Response, "Internal Server Error", fasthttp.StatusInternalServerError)
			break
		}
	}

	if currentLayer == len(s.layers) {
		currentLayer--
		s.resetHeaders(&ctx.Request.Header, requestHeaders)
		ctx.Request.Header.SetMethodBytes(methodBytes)
		ctx.Request.Header.SetRequestURI(uri)

		layerTracer.StartExecute()
		s.executor(state)
		layerTracer.FinishExecute()
	}
	if err2 := ParseHeaders(responseHeaders, state.Response.Header.Header()); err2 != nil {
		s.logger.Debug("[%s] (%d) %s %s: malformed response headers: %s",
			ctx.RemoteAddr(), reqID, method, uri, err)
		MakeSimpleResponse(&ctx.Response, "Malformed response headers", fasthttp.StatusBadRequest)
		return
	}

	for ; currentLayer >= 0; currentLayer-- {
		layerTracer.StartOnResponse()
		s.layers[currentLayer].OnResponse(state, err)
		layerTracer.FinishOnResponse()
	}

	responseCode := ctx.Response.Header.StatusCode()
	s.resetHeaders(&ctx.Response.Header, responseHeaders)
	ctx.Response.SetStatusCode(responseCode)
}

func (s *Server) resetHeaders(headers fasthttpHeader, set *HeaderSet) {
	contentLength := headers.ContentLength()
	headers.Reset()
	headers.DisableNormalizing()
	for _, v := range set.Items() {
		headers.SetBytesKV(v.Key, v.Value)
	}
	headers.SetContentLength(contentLength)
}

// NewServer creates new instance of the Server.
func NewServer(opts ServerOpts) (*Server, error) {
	metrics := opts.GetMetrics()
	certs, err := ca.NewCA(opts.GetCertCA(),
		opts.GetCertKey(),
		metrics,
		opts.GetTLSCertCacheSize(),
		opts.GetTLSCertCachePrune(),
		opts.GetOrganizationName())
	if err != nil {
		return nil, errors.Annotate(err, "Cannot create CA")
	}

	srv := &Server{
		certs:      certs,
		executor:   opts.GetExecutor(),
		layers:     opts.GetLayers(),
		logger:     opts.GetLogger(),
		metrics:    metrics,
		tracerPool: opts.GetTracerPool(),
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
