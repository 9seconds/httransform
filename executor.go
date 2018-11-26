package main

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

var DefaultHTTPClient fasthttp.Client

type Executor func(*LayerState)

func ProxyExecutor(state *LayerState) {
	err := DefaultHTTPClient.Do(state.Request, state.Response)
	if err != nil {
		resp := fasthttp.AcquireResponse()
		MakeBadResponse(resp, fmt.Sprintf("Cannot fetch from upstream: %s", err), fasthttp.StatusBadGateway)
		resp.CopyTo(state.Response)
		fasthttp.ReleaseResponse(resp)
	}
}

func init() {
	DefaultHTTPClient = fasthttp.Client{
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
	}
}
