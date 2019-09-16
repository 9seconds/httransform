package httransform

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"

	"github.com/9seconds/httransform/client"
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

var (
	defaultTLSConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gosec
)

// HTTPRequestExecutor is an interface to be used for ExecuteRequest and
// ExecuteRequestTimeout functions.
type HTTPRequestExecutor interface {
	Do(*fasthttp.Request, *fasthttp.Response) error
	DoTimeout(*fasthttp.Request, *fasthttp.Response, time.Duration) error
}

// MakeStreamingClosingHTTPClient returns HTTPRequestExecutor which
// streams body response but closes connection after every request.
func MakeStreamingClosingHTTPClient() HTTPRequestExecutor {
	return makeStreamingClosingHTTPClient(client.FastHTTPBaseDialer)
}

// MakeStreamingClosingSOCKS5HTTPClient returns HTTPRequestExecutor
// which streams body response but closes connection after every
// request.
//
// This client uses given SOCKS5 proxy.
func MakeStreamingClosingSOCKS5HTTPClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeStreamingClosingHTTPClient(makeSOCKS5BaseDialer(proxyURL))
}

// MakeStreamingClosingProxyHTTPClient returns HTTPRequestExecutor which
// streams body response but closes connection after every request.
//
// This client uses given HTTP proxy. You should not use this client for
// HTTPS requests.
func MakeStreamingClosingProxyHTTPClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeStreamingClosingHTTPClient(makeProxyBaseDialer(proxyURL))
}

// MakeStreamingClosingCONNECTHTTPClient returns HTTPRequestExecutor
// which streams body response but closes connection after every
// request.
//
// This client uses given HTTPS proxy. You should not use this client
// for HTTP requests.
func MakeStreamingClosingCONNECTHTTPClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeStreamingClosingHTTPClient(makeCONNECTBaseDialer(proxyURL, client.FastHTTPBaseDialer))
}

// MakeStreamingReuseHTTPClient returns HTTPRequestExecutor which
// streams body response and maintains a pool of connections to the
// hosts.
func MakeStreamingReuseHTTPClient() HTTPRequestExecutor {
	return makeStreamingPooledHTTPClient(client.FastHTTPBaseDialer)
}

// MakeStreamingReuseSOCKS5HTTPClient returns HTTPRequestExecutor which
// streams body response and maintains a pool of connections to the
// hosts.
//
// This client uses given SOCKS5 proxy.
func MakeStreamingReuseSOCKS5HTTPClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeStreamingPooledHTTPClient(makeSOCKS5BaseDialer(proxyURL))
}

// MakeStreamingReuseProxyHTTPClient returns HTTPRequestExecutor which
// streams body response and maintains a pool of connections to the
// hosts.
//
// This client uses given HTTP proxy. You should not use this client for
// HTTPS requests.
func MakeStreamingReuseProxyHTTPClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeStreamingPooledHTTPClient(makeProxyBaseDialer(proxyURL))
}

// MakeStreamingReuseCONNECTHTTPClient returns HTTPRequestExecutor which
// streams body response and maintains a pool of connections to the
// hosts.
//
// This client uses given HTTPS proxy. You should not use this client
// for HTTP requests.
func MakeStreamingReuseCONNECTHTTPClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeStreamingPooledHTTPClient(makeCONNECTBaseDialer(proxyURL, client.FastHTTPBaseDialer))
}

// MakeDefaultHTTPClient returns HTTPRequestExecutor which fetches
// the whole body in memory.
//
// Please use this client if you need production-grade HTTP client.
func MakeDefaultHTTPClient() HTTPRequestExecutor {
	return makeDefaultHTTPClient(fasthttp.DialDualStack)
}

// MakeDefaultSOCKS5ProxyClient returns HTTPRequestExecutor which fetches
// the whole body in memory.
//
// Please use this client if you need production-grade HTTP client.
//
// This client uses given SOCKS5 proxy.
func MakeDefaultSOCKS5ProxyClient(proxyURL *url.URL) HTTPRequestExecutor {
	base := makeSOCKS5BaseDialer(proxyURL)
	return makeDefaultHTTPClient(func(addr string) (net.Conn, error) {
		return base(addr, ConnectDialTimeout)
	})
}

// MakeDefaultHTTPProxyClient returns HTTPRequestExecutor which fetches
// the whole body in memory.
//
// Please use this client if you need production-grade HTTP client.
//
// This client uses given HTTP proxy. You should not use this client for
// HTTPS requests.
func MakeDefaultHTTPProxyClient(proxyURL *url.URL) HTTPRequestExecutor {
	base := makeProxyBaseDialer(proxyURL)
	return makeDefaultHTTPClient(func(addr string) (net.Conn, error) {
		return base(addr, ConnectDialTimeout)
	})
}

// MakeDefaultCONNECTProxyClient returns HTTPRequestExecutor which
// fetches the whole body in memory.
//
// Please use this client if you need production-grade HTTP client.
//
// This client uses given HTTPS proxy. You should not use this client
// for HTTP requests.
func MakeDefaultCONNECTProxyClient(proxyURL *url.URL) HTTPRequestExecutor {
	base := makeCONNECTBaseDialer(proxyURL, client.FastHTTPBaseDialer)
	return makeDefaultHTTPClient(func(addr string) (net.Conn, error) {
		return base(addr, ConnectDialTimeout)
	})
}

func makeStreamingClosingHTTPClient(dialer client.BaseDialer) HTTPRequestExecutor {
	newDialer, _ := client.NewSimpleDialer(dialer, ConnectDialTimeout)
	return client.NewClient(newDialer, defaultTLSConfig)
}

func makeStreamingPooledHTTPClient(dialer client.BaseDialer) HTTPRequestExecutor {
	newDialer, _ := client.NewPooledDialer(dialer, ConnectDialTimeout, MaxConnsPerHost)
	return client.NewClient(newDialer, defaultTLSConfig)
}

func makeDefaultHTTPClient(dialFunc fasthttp.DialFunc) *fasthttp.Client {
	return &fasthttp.Client{
		Dial:                          dialFunc,
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
		MaxConnsPerHost:               MaxConnsPerHost,
		TLSConfig:                     defaultTLSConfig,
	}
}

func makeSOCKS5BaseDialer(proxyURL *url.URL) client.BaseDialer {
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
		panic(err)
	}

	return func(addr string, _ time.Duration) (net.Conn, error) {
		return proxyDialer.Dial("tcp", addr)
	}
}

func makeProxyBaseDialer(proxyURL *url.URL) client.BaseDialer {
	return func(_ string, timeout time.Duration) (net.Conn, error) {
		return client.FastHTTPBaseDialer(proxyURL.Host, timeout)
	}
}

func makeCONNECTBaseDialer(proxyURL *url.URL, base client.BaseDialer) client.BaseDialer {
	proxyAuthHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)

	return func(addr string, timeout time.Duration) (net.Conn, error) {
		conn, err := base(proxyURL.Host, timeout)
		if err != nil {
			return nil, xerrors.Errorf("cannot dial to proxy %s: %w", proxyURL.Host, err)
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
			return nil, xerrors.Errorf("cannot send CONNECT request: %w", err)
		}

		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)

		res.SkipBody = true
		if err := res.Read(bufio.NewReader(conn)); err != nil {
			conn.Close()
			return nil, xerrors.Errorf("cannot read from proxy after CONNECT request: %w", err)
		}

		if res.Header.StatusCode() != 200 {
			conn.Close()
			return nil, xerrors.Errorf("Proxy responded %d on CONNECT request", res.Header.StatusCode())
		}

		return conn, nil
	}
}
