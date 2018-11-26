package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/ca"
)

type Server struct {
	serverPool sync.Pool
	server     *fasthttp.Server
	certs      ca.CA
	layers     []Layer
	executor   Executor
}

func (s *Server) initialHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.RequestURI())

	if ctx.IsConnect() {
		host, port, err := SplitHostPort(path)
		if err != nil {
			s.defaultErrorHandler(ctx, errors.Annotatef(err, "Cannot split host/port for path %s", path))
			return
		}

		handler, err := s.makeHijackHandler(host, port)
		if err != nil {
			s.defaultErrorHandler(ctx, errors.Annotate(err, "Cannot create hijack handler"))
			return
		}
		ctx.Hijack(handler)
		ctx.Success("", nil)

		return
	}

	host, port, err := HostPortFromURL(path)
	if err != nil {
		s.defaultErrorHandler(ctx, errors.Annotatef(err, "Cannot parse host/port for URL %s", path))
	}

	ctx.SetUserValue("host", host)
	ctx.SetUserValue("port", port)

	s.handleRequest(ctx)
}

func (s *Server) makeHijackHandler(host string, port int) (fasthttp.HijackHandler, error) {
	return func(conn net.Conn) {
		defer conn.Close()

		conf, err := s.certs.Get(host)
		if err != nil {
			// TODO
			return
		}
		defer conf.Release()

		tlsConn := tls.Server(conn, conf.Get())
		defer tlsConn.Close()

		if err = tlsConn.Handshake(); err != nil {
			// TODO
			return
		}

		srv := s.serverPool.Get().(*fasthttp.Server)
		defer s.serverPool.Put(srv)

		srv.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.SetUserValue("host", host)
			ctx.SetUserValue("port", port)
			s.handleRequest(ctx)
		}
		srv.ServeConn(tlsConn)
	}, nil
}

func (s *Server) defaultErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	ctx.Error("Incorrect request", fasthttp.StatusBadRequest)
}

func (s *Server) handleRequest(ctx *fasthttp.RequestCtx) {
	requestHeaders := getHeaderSet()
	responseHeaders := getHeaderSet()
	defer func() {
		releaseHeaderSet(requestHeaders)
		releaseHeaderSet(responseHeaders)
	}()

	if err := parseHeaders(requestHeaders, ctx.Request.Header.Header()); err != nil {
		ctx.Error("Malformed request", fasthttp.StatusBadRequest)
		return
	}

	state := getLayerState()
	defer releaseLayerState(state)
	initLayerState(state, ctx, requestHeaders, responseHeaders)

	currentLayer := -1
	requestMethod := ctx.Request.Header.Method()
	requestURI := ctx.Request.Header.RequestURI()
	for _, layer := range s.layers {
		if err := layer.OnRequest(state); err != nil {
			// TODO MANAGE ERROR
			break
		}
		currentLayer++
	}
	ctx.Request.Header.Reset()
	ctx.Request.Header.DisableNormalizing()
	for _, v := range requestHeaders.Items() {
		ctx.Request.Header.SetBytesKV(v.Key, v.Value)
	}
	ctx.Request.Header.SetMethodBytes(requestMethod)
	ctx.Request.Header.SetRequestURIBytes(requestURI)

	if err := s.executor(state); err != nil {
		// TODO
		fmt.Println(err)
		ctx.Error("Malformed response", fasthttp.StatusBadRequest)
		return
	}

	// TODO PERFORM REQUEST
	if err := parseHeaders(responseHeaders, state.Response.Header.Header()); err != nil {
		ctx.Error("Malformed response", fasthttp.StatusBadRequest)
		return
	}
	responseCode := ctx.Response.Header.StatusCode()
	for currentLayer >= 0 {
		if err := s.layers[currentLayer].OnResponse(state); err != nil {
			// TODO MANAGE ERROR
			break
		}
		currentLayer--
	}

	ctx.Response.Header.Reset()
	ctx.Response.Header.DisableNormalizing()
	for _, v := range responseHeaders.Items() {
		ctx.Response.Header.SetBytesKV(v.Key, v.Value)
	}
	ctx.Response.SetStatusCode(responseCode)
}

func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown()
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
		serverPool: sync.Pool{
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
		},
	}
	srv.server = srv.serverPool.Get().(*fasthttp.Server)
	srv.server.Handler = srv.initialHandler

	return srv, nil
}

var DefaultCertCA = []byte(`-----BEGIN CERTIFICATE-----
MIIEjDCCAvGgAwIBAgIJALFJCQU3CJNhMA0GCSqGSIb3DQEBCwUAMFoxCzAJBgNV
BAYTAklSMRMwEQYDVQQIDApTb21lLVN0YXRlMRQwEgYDVQQKDAtTY3JhcGluZ0h1
YjEgMB4GA1UECwwXQ3Jhd2xlcmEgSGVhZGxlc3MgUHJveHkwHhcNMTgwNDA1MTAz
MjI1WhcNNDUwODIxMTAzMjI1WjBaMQswCQYDVQQGEwJJUjETMBEGA1UECAwKU29t
ZS1TdGF0ZTEUMBIGA1UECgwLU2NyYXBpbmdIdWIxIDAeBgNVBAsMF0NyYXdsZXJh
IEhlYWRsZXNzIFByb3h5MIIBpDANBgkqhkiG9w0BAQEFAAOCAZEAMIIBjAKCAYMM
Ni6ZLBoLN6Ut+amWbI2JHN1jnkrh25HCA9HkdboEj/oN+O8xcKV0UsVBwElCz20B
lXbUbEwmjn8a93LfTLUT0uM8Zt1AOS/kMQ1mmJ3ZDe9DWxgtbt69YUtE0RGCy5IY
QQPmDcwvZ8EE0PHARXKNiSNJbu7FrgtWZUVT/ND2SrFkO+PQJeobGuooaEROk3Hu
Qkqqe8w063XmMhIovOnDi4FBNbxDd9n6wSV4ngyjOyqunJOJ9hy1TZweNbZnGvSi
keL30YOHuhy6uuoxLiCnD0QlUY37WwF884/Ozxk6InL29Fo8x5JzC3+cj7bT4hr7
c9kW5Te3qxezYiGgE+I/7VgGZhkJY+Ff6SD9NncB7QiiKDrqUvUu5ZKtFdSRi5gL
PX8UX6nfBEny6fm9/7NQ8FcebneLe4nLGy3iLzcUvOdS8wwbav8kutBHEGEyzFkn
6sxu1sNWtyS4W/vUlaELszaBEjinV+d5Ittjqz1Otl84+tk2/O5Rdzp/dsnEe2vu
GBMCAwEAAaNQME4wHQYDVR0OBBYEFHGYnNHKY/cTwRD1r5nqIXxNxQrKMB8GA1Ud
IwQYMBaAFHGYnNHKY/cTwRD1r5nqIXxNxQrKMAwGA1UdEwQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggGEAAV4cqweWxi5PvQ44PnGs+9DG7Nj7mXcdgl8f7zu1/vGEPOz
FD4Pd3bjAwsYg92Wd5SwjytTfAjFvgVPV6NBrUlHUaD7Bvge6mKBqY/LkGzises9
NE8tMSjFhAIrvhtgf9LdiZ1vZeKBBkRL/ZE8eV6tvmZ643TeW9xVog2dR3GPWTJy
nkOKaBzAf/gT3yZHFnOtZfUbBVGydy/YmqBt3Nc6RKxfSr++nDYySTxV2T1U0jcb
nehzLgk3du96BiVmEyuBcfD1tlaY6SNHyaGIzBGdKf4QDULfCDr0px2gWeRkWS7w
sB81yI6yLxRr+kN/zHX1oDV0ufChVHlbYHLID6wK22zN4CWwJNXXPmPXoiG9kF/Z
CXGbAJCTiER5vO40eyY1P72NOyLyqSd6KvGW1xYAlGLEhVcry8SJ6DXun2hwFV7y
WxuT0PYISCxxy6APrg+WdjQRMbiMxP9dNRmHnbCPzYO4A1z8Iuvz5AdJltItAjJu
ZYFeyU9HcskJWkXUCISLDQ==
-----END CERTIFICATE-----`)

// DefaultPrivateKey is hardcoded TLS private key
// nolint: gochecknoglobals
var DefaultPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
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
2Mu+
-----END RSA PRIVATE KEY-----`)

func main() {
	ln, _ := net.Listen("tcp", "127.0.0.1:3128")
	srv, _ := NewServer(ServerOpts{
		CertCA:           DefaultCertCA,
		CertKey:          DefaultPrivateKey,
		OrganizationName: "TEST",
	}, []Layer{}, ProxyExecutor)
	srv.Serve(ln)
}
