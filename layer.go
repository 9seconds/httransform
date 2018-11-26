package main

import "github.com/valyala/fasthttp"

type LayerState struct {
	RequestID       uint64
	Host            string
	Port            int
	RequestHeaders  *HeaderSet
	ResponseHeaders *HeaderSet
	Request         *fasthttp.Request
	Response        *fasthttp.Response
}

type Layer interface {
	OnRequest(*LayerState) error
	OnResponse(*LayerState)
}

type BaseLayer struct {
}

func (b BaseLayer) OnRequest(_ *LayerState) error {
	return nil
}

func (b BaseLayer) OnResponse(_ *LayerState) {
}

func initLayerState(state *LayerState, ctx *fasthttp.RequestCtx,
	requestHeaders, responseHeaders *HeaderSet) {
	state.RequestID = ctx.ID()
	state.RequestHeaders = requestHeaders
	state.ResponseHeaders = responseHeaders
	state.Request = &ctx.Request
	state.Response = &ctx.Response
	state.Host = ctx.UserValue("host").(string)
	state.Port = ctx.UserValue("port").(int)
}
