package httransform

import (
	"bytes"
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

		return func(state *LayerState) {
			if len(proxyAuthorizationHeaderValue) > 0 {
				state.Request.Header.SetBytesV("Proxy-Authorization", proxyAuthorizationHeaderValue)
			}

			if !state.isConnect {
				host := state.Request.Host()
				if bytes.IndexByte(host, ':') < 0 {
					state.Request.Header.SetHostBytes(append(host, ':', '8', '0'))
				}

				uri := fasthttp.AcquireURI()
				defer fasthttp.ReleaseURI(uri)

				uri.Parse(nil, state.Request.RequestURI())
				state.Request.SetRequestURIBytes(uri.RequestURI())
			}

			ExecuteRequest(client, state.Request, state.Response)
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
