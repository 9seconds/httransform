package httransform

import (
	"fmt"
	"net/url"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

var executorDefaultHTTPClient = MakeHTTPClient()

// Executor presents a type which converts HTTP request to HTTP
// response. It is not necessary that executor is limited to HTTP. It
// can do anything, but as a result, it has to update response.
type Executor func(*LayerState)

// HTTPExecutor is the default executor and the simplies one you want
// to use. It does takes HTTP request from the state and execute the
// request, returning response after that.
func HTTPExecutor(state *LayerState) {
	ExecuteRequestTimeout(executorDefaultHTTPClient, state.Request, state.Response, DefaultHTTPTImeout)
}

// MakeProxyChainExecutor is the factory which produce executors
// which execute given HTTP request using HTTP proxy. In other words,
// MakeProxyChainExecutor allows to chain HTTP proxies. This executor
// supports SOCKS5, HTTP and HTTPS proxies.
func MakeProxyChainExecutor(proxyURL *url.URL) (Executor, error) {
	switch proxyURL.Scheme {
	case "socks5":
		client, err := MakeProxySOCKS5Client(proxyURL)
		if err != nil {
			return nil, err
		}

		return func(state *LayerState) {
			ExecuteRequestTimeout(client, state.Request, state.Response, DefaultHTTPTImeout)
		}, nil

	case "http", "https", "":
		httpProxyClient := MakeHTTPProxyClient(proxyURL)
		httpsProxyClient := MakeHTTPSProxyClient(proxyURL)
		proxyAuthorizationHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)

		return func(state *LayerState) {
			if len(proxyAuthorizationHeaderValue) > 0 {
				state.Request.Header.SetBytesV("Proxy-Authorization", proxyAuthorizationHeaderValue)
			}

			client := httpsProxyClient
			if !state.isConnect {
				client = httpProxyClient
			}

			ExecuteRequestTimeout(client, state.Request, state.Response, DefaultHTTPTImeout)
		}, nil
	}

	return nil, errors.Errorf("Unknown proxy URL scheme %s", proxyURL.Scheme)
}

// ExecuteRequest is a basic method on executing requests. On error, it
// also converts response to default 500.
func ExecuteRequest(client HTTPRequestExecutor, req *fasthttp.Request, resp *fasthttp.Response) {
	if err := client.Do(req, resp); err != nil {
		failResponse(resp, err)
	}
}

// ExecuteRequestTimeout does the same as ExecuteRequest but also
// uses given timeout.
func ExecuteRequestTimeout(client HTTPRequestExecutor, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) {
	if err := client.DoTimeout(req, resp, timeout); err != nil {
		failResponse(resp, err)
	}
}

func failResponse(resp *fasthttp.Response, err error) {
	newResponse := fasthttp.AcquireResponse()
	MakeSimpleResponse(newResponse, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
	newResponse.CopyTo(resp)
	fasthttp.ReleaseResponse(newResponse)
}
