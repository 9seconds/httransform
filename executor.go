package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/proxy"
)

type Executor func(*LayerState)

var (
	DefaultHTTPClient *fasthttp.Client
	HTTPExecutor      Executor
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
		proxyClient.Dial = makeHTTPProxyDialer(proxyURL)

		proxyAuthorizationHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)
		return func(state *LayerState) {
			if proxyAuthorizationHeaderValue != "" {
				state.ResponseHeaders.SetString("Proxy-Authorization", proxyAuthorizationHeaderValue)
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

func MakeProxyAuthorizationHeaderValue(proxyURL *url.URL) string {
	username := proxyURL.User.Username()
	password, ok := proxyURL.User.Password()
	if ok || username != "" {
		line := fmt.Sprintf("%s:%s", username, password)
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(line))
	}

	return ""
}

func MakeHTTPClient() *fasthttp.Client {
	return &fasthttp.Client{
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
	}
}

func executeRequest(client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) {
	if err := client.Do(req, resp); err != nil {
		newResponse := fasthttp.AcquireResponse()
		MakeBadResponse(resp, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
		resp.CopyTo(resp)
		fasthttp.ReleaseResponse(newResponse)
	}
}

func makeHTTPProxyDialer(proxyURL *url.URL) fasthttp.DialFunc {
	proxyAuthHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)

	return func(addr string) (net.Conn, error) {
		conn, err := fasthttp.Dial(proxyURL.Host)
		if err != nil {
			return nil, errors.Annotatef(err, "Cannot dial to proxy %s", proxyURL.Host)
		}

		connectRequest := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
		if proxyAuthHeaderValue != "" {
			connectRequest += fmt.Sprintf("Proxy-Authorization: %s\r\n", proxyAuthHeaderValue)
		}
		connectRequest += "\r\n"

		if _, err := conn.Write([]byte(connectRequest)); err != nil {
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
}
