package main

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

var DefaultHTTPClient *fasthttp.Client

type Executor func(*LayerState)

func ProxyExecutor(state *LayerState) {
	err := DefaultHTTPClient.Do(state.Request, state.Response)
	if err != nil {
		resp := fasthttp.AcquireResponse()
		resp.SetBodyString(fmt.Sprintf("Cannot fetch from upstream: %s", err))
		resp.SetStatusCode(fasthttp.StatusBadGateway)
		resp.Header.SetContentType("text/plain")
		resp.CopyTo(state.Response)
		fasthttp.ReleaseResponse(resp)
	}
}

func init() {
	DefaultHTTPClient = &fasthttp.Client{
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
	}
}
