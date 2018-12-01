package httransform

import (
	"fmt"
	"net/url"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

var executorDefaultHTTPClient HTTPRequestExecutor

type Executor func(*LayerState)

func HTTPExecutor(state *LayerState) {
	ExecuteRequestTimeout(executorDefaultHTTPClient, state.Request, state.Response, DefaultHTTPTImeout)
}

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

func ExecuteRequest(client HTTPRequestExecutor, req *fasthttp.Request, resp *fasthttp.Response) {
	if err := client.Do(req, resp); err != nil {
		newResponse := fasthttp.AcquireResponse()
		MakeSimpleResponse(resp, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
		resp.CopyTo(resp)
		fasthttp.ReleaseResponse(newResponse)
	}
}

func ExecuteRequestTimeout(client HTTPRequestExecutor, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) {
	if err := client.DoTimeout(req, resp, timeout); err != nil {
		newResponse := fasthttp.AcquireResponse()
		MakeSimpleResponse(resp, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
		resp.CopyTo(resp)
		fasthttp.ReleaseResponse(newResponse)
	}
}

func init() {
	executorDefaultHTTPClient = MakeHTTPClient()
}
