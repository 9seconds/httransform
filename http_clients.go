package httransform

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
)

// Defaults for HTTP clients.
const (
	// MaxConnsPerHost is the maximal number of open connections to some
	// host in HTTP client.
	MaxConnsPerHost = 8192

	// ConnectDialTimeout is the timeout to establish TCP connection to the
	// host. Usually it is used only for proxy chains.
	ConnectDialTimeout = 30 * time.Second

	// DefaultHTTPTImeout is the timeout from starting of HTTP request to
	// getting a response.
	DefaultHTTPTImeout = 3 * time.Minute
)

// HTTP is the default HTTPRequestExecutor intended to be used by the
// user. Default executors use other client.
var HTTP = MakeHTTPClient()

// HTTPRequestExecutor is an interface to be used for ExecuteRequest and
// ExecuteRequestTimeout functions.
type HTTPRequestExecutor interface {
	Do(*fasthttp.Request, *fasthttp.Response) error
	DoTimeout(*fasthttp.Request, *fasthttp.Response, time.Duration) error
}

// MakeHTTPClient creates a new instance of HTTPRequestExecutor. This
// executor knows nothing about proxies, simply executes given HTTP
// request and writes to HTTP response.
func MakeHTTPClient() HTTPRequestExecutor {
	return makeHTTPClient()
}

// MakeProxySOCKS5Client creates a new instance of HTTPRequestExecutor
// which uses given SOCKS5 proxy from URL.
func MakeProxySOCKS5Client(proxyURL *url.URL) (HTTPRequestExecutor, error) {
	client := makeHTTPClient()
	dialer, err := makeSOCKS5Dialer(proxyURL)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot build SOCKS5 client")
	}

	client.Dial = dialer

	return client, nil
}

// MakeHTTPProxyClient creates a new instance of HTTPRequestExecutor
// which can work with HTTP proxies. Please pay attention that you can
// execute only HTTP requests using this executor, it does not support
// request of HTTPS resources.
func MakeHTTPProxyClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeHTTPHostClient(proxyURL.Host)
}

// MakeHTTPSProxyClient creates a new instance of HTTPRequestExecutor
// which can work with HTTPS proxies (i.e execute CONNECT request
// first). It does not work with plain HTTP requests.
func MakeHTTPSProxyClient(proxyURL *url.URL) HTTPRequestExecutor {
	client := makeHTTPClient()
	client.Dial = makeHTTPProxyDialer(proxyURL)

	return client
}

func makeSOCKS5Dialer(proxyURL *url.URL) (fasthttp.DialFunc, error) {
	var auth *proxy.Auth
	username := proxyURL.User.Username()
	password, ok := proxyURL.User.Password()
	if ok || username != "" {
		auth = &proxy.Auth{
			User:     username,
			Password: password,
		}
	}

	proxyDialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot build socks5 proxy dialer")
	}

	return func(addr string) (net.Conn, error) {
		return proxyDialer.Dial("tcp", addr)
	}, nil
}

func makeHTTPProxyDialer(proxyURL *url.URL) fasthttp.DialFunc {
	proxyAuthHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)

	return func(addr string) (net.Conn, error) {
		conn, err := fasthttp.DialDualStackTimeout(proxyURL.Host, ConnectDialTimeout)
		if err != nil {
			return nil, errors.Annotatef(err, "Cannot dial to proxy %s", proxyURL.Host)
		}
		if _, _, err = net.SplitHostPort(addr); err != nil {
			addr = net.JoinHostPort(addr, "443")
		}

		connectRequest := connectRequestBufferPool.Get().(*bytes.Buffer)
		connectRequest.Reset()
		defer connectRequestBufferPool.Put(connectRequest)

		connectRequest.Write([]byte("CONNECT "))
		connectRequest.WriteString(addr)
		connectRequest.Write([]byte(" HTTP/1.1\r\nHost: "))
		connectRequest.WriteString(addr)
		connectRequest.Write([]byte("\r\n"))

		if proxyAuthHeaderValue != nil {
			connectRequest.Write([]byte("Proxy-Authorization: "))
			connectRequest.Write(proxyAuthHeaderValue)
			connectRequest.Write([]byte("\r\n"))
		}
		connectRequest.Write([]byte("\r\n"))

		if _, err := conn.Write(connectRequest.Bytes()); err != nil {
			return nil, errors.Annotate(err, "Cannot send CONNECT request")
		}

		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)

		res.SkipBody = true
		if err := res.Read(bufio.NewReader(conn)); err != nil {
			conn.Close()
			return nil, errors.Annotate(err, "Cannot read from proxy after CONNECT request")
		}

		if res.Header.StatusCode() != 200 {
			conn.Close()
			return nil, errors.Errorf("Proxy responded %d on CONNECT request", res.Header.StatusCode())
		}

		return conn, nil
	}
}

func makeHTTPClient() *fasthttp.Client {
	return &fasthttp.Client{
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
		MaxConnsPerHost:               MaxConnsPerHost,
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true}, // nolint: gosec
	}
}

func makeHTTPHostClient(addr string) *fasthttp.HostClient {
	return &fasthttp.HostClient{
		Addr:                          addr,
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
		MaxConns:                      MaxConnsPerHost,
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true}, // nolint: gosec
	}
}
