package main

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

var DefaultHTTPClient *fasthttp.Client

type Executor func(*LayerState) error

func ProxyExecutor(state *LayerState) error {
	fmt.Println(string(state.Request.Header.Method()))
	fmt.Println(string(state.Request.RequestURI()))
	fmt.Println(string(state.Request.Host()))
	fmt.Println(string(state.Request.String()))
	return DefaultHTTPClient.Do(state.Request, state.Response)
}

func init() {
	DefaultHTTPClient = &fasthttp.Client{
		DialDualStack:                 true,
		DisableHeaderNamesNormalizing: true,
	}
}
