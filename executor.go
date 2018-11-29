package httransform

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
)

type Executor func(*LayerState)

var (
	DefaultHTTPClient *fasthttp.Client
	HTTPExecutor      Executor
	bytesBufferPool   sync.Pool
)

func MakeHTTPExecutor(client *fasthttp.Client) Executor {
	return func(state *LayerState) {
		executeRequest(client, state.Request, state.Response)
	}
}

func ProxyChainExecutor(proxyURL *url.URL) (Executor, error) {
	proxyClient := MakeHTTPClient()

	switch proxyURL.Scheme {
	case "socks5":
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

		proxyClient.Dial = func(addr string) (net.Conn, error) {
			return proxyDialer.Dial("tcp", addr)
		}

		return MakeHTTPExecutor(proxyClient), nil

	case "", "http", "https":
		proxyClient.Dial = MakeHTTPProxyDialer(proxyURL)

		proxyAuthorizationHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)
		return func(state *LayerState) {
			if proxyAuthorizationHeaderValue != nil {
				state.ResponseHeaders.SetBytes([]byte("Proxy-Authorization"), proxyAuthorizationHeaderValue)
			}
			if state.isConnect {
				executeRequest(proxyClient, state.Request, state.Response)
			} else {
				executeRequest(DefaultHTTPClient, state.Request, state.Response)
			}
		}, nil
	}

	return nil, errors.Errorf("Unknown proxy scheme %s", proxyURL.Scheme)
}

func MakeProxyAuthorizationHeaderValue(proxyURL *url.URL) []byte {
	username := proxyURL.User.Username()
	password, ok := proxyURL.User.Password()
	if ok || username != "" {
		line := username + ":" + password
		return []byte("Basic " + base64.StdEncoding.EncodeToString([]byte(line)))
	}

	return nil
}

func MakeHTTPClient() *fasthttp.Client {
	return &fasthttp.Client{
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
	}
}

func executeRequest(client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) {
	if err := client.Do(req, resp); err != nil {
		newResponse := fasthttp.AcquireResponse()
		MakeSimpleResponse(resp, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
		resp.CopyTo(resp)
		fasthttp.ReleaseResponse(newResponse)
	}
}

func MakeHTTPProxyDialer(proxyURL *url.URL) fasthttp.DialFunc {
	proxyAuthHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)

	return func(addr string) (net.Conn, error) {
		conn, err := fasthttp.Dial(proxyURL.Host)
		if err != nil {
			return nil, errors.Annotatef(err, "Cannot dial to proxy %s", proxyURL.Host)
		}

		connectRequest := bytesBufferPool.Get().(*bytes.Buffer)
		connectRequest.Reset()
		defer bytesBufferPool.Put(connectRequest)

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

func init() {
	DefaultHTTPClient = MakeHTTPClient()
	HTTPExecutor = MakeHTTPExecutor(DefaultHTTPClient)
	bytesBufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
}
