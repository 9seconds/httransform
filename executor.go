package httransform

import (
	"fmt"
	"net/url"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

var executorDefaultHttpClient *fasthttp.Client

type Executor func(*LayerState)

func HTTPExecutor(state *LayerState) {
	ExecuteRequest(executorDefaultHttpClient, state.Request, state.Response)
}

func MakeProxyChainExecutor(proxyURL *url.URL) (Executor, error) {
	switch proxyURL.Scheme {
	case "socks5":
		client, err := MakeProxySOCKS5Client(proxyURL)
		if err != nil {
			return nil, err
		}

		return func(state *LayerState) {
			ExecuteRequest(client, state.Request, state.Response)
		}, nil

	case "http", "https", "":
		client := MakeHTTPProxyClient(proxyURL)
		proxyAuthorizationHeaderValue := MakeProxyAuthorizationHeaderValue(proxyURL)
		if len(proxyAuthorizationHeaderValue) == 0 {
			return func(state *LayerState) {
				if state.isConnect {
					ExecuteRequest(client, state.Request, state.Response)
				} else {
					ExecuteRequest(executorDefaultHttpClient, state.Request, state.Response)
				}
			}, nil
		}

		return func(state *LayerState) {
			state.Request.Header.SetBytesV("Proxy-Authorization", proxyAuthorizationHeaderValue)
			if state.isConnect {
				ExecuteRequest(client, state.Request, state.Response)
			} else {
				ExecuteRequest(executorDefaultHttpClient, state.Request, state.Response)
			}
		}, nil
	}

	return nil, errors.Errorf("Unknown proxy URL scheme %s", proxyURL.Scheme)
}

func ExecuteRequest(client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) {
	if err := client.Do(req, resp); err != nil {
		newResponse := fasthttp.AcquireResponse()
		MakeSimpleResponse(resp, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
		resp.CopyTo(resp)
		fasthttp.ReleaseResponse(newResponse)
	}
}

func init() {
	executorDefaultHttpClient = MakeHTTPClient()
}
