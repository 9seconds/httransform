package httransform

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"net"
	"net/url"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
)

const MaxConnsPerHost = 8192

var HTTP HTTPRequestExecutor

type HTTPRequestExecutor interface {
	Do(*fasthttp.Request, *fasthttp.Response) error
}

func MakeHTTPClient() HTTPRequestExecutor {
	return makeHTTPClient()
}

func MakeProxySOCKS5Client(proxyURL *url.URL) (HTTPRequestExecutor, error) {
	client := makeHTTPClient()
	dialer, err := makeSOCKS5Dialer(proxyURL)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot build SOCKS5 client")
	}

	client.Dial = dialer

	return client, nil
}

func MakeHTTPProxyClient(proxyURL *url.URL) HTTPRequestExecutor {
	return makeHTTPHostClient(proxyURL.Host)
}

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
		conn, err := fasthttp.Dial(proxyURL.Host)
		if err != nil {
			return nil, errors.Annotatef(err, "Cannot dial to proxy %s", proxyURL.Host)
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
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
	}
}

func makeHTTPHostClient(addr string) *fasthttp.HostClient {
	return &fasthttp.HostClient{
		Addr:                          addr,
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
		MaxConns:                      MaxConnsPerHost,
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
	}
}

func init() {
	HTTP = MakeHTTPClient()
}
