package main

import "github.com/valyala/fasthttp"

type Layer interface {
	OnRequest(*LayerState) error
	OnResponse(*LayerState)
}

type LayerState struct {
	isConnect bool
	ctx       map[string]interface{}

	RequestID       uint64
	RequestHeaders  *HeaderSet
	ResponseHeaders *HeaderSet
	Request         *fasthttp.Request
	Response        *fasthttp.Response
}

func (l *LayerState) Set(key string, value interface{}) {
	l.ctx[key] = value
}

func (l *LayerState) Get(key string) (interface{}, bool) {
	item, ok := l.ctx[key]
	return item, ok
}

type BaseLayer struct {
}

func (b BaseLayer) OnRequest(_ *LayerState) error {
	return nil
}

func (b BaseLayer) OnResponse(_ *LayerState) {
}

func initLayerState(state *LayerState, ctx *fasthttp.RequestCtx,
	requestHeaders, responseHeaders *HeaderSet,
	isConnect bool) {
	state.RequestID = ctx.ID()
	state.RequestHeaders = requestHeaders
	state.ResponseHeaders = responseHeaders
	state.Request = &ctx.Request
	state.Response = &ctx.Response
	state.isConnect = isConnect
}

type ConnectionCloseLayer struct {
	BaseLayer
}

func (c ConnectionCloseLayer) OnResponse(state *LayerState) {
	state.ResponseHeaders.SetString("Connection", "close")
}

type ProxyHeadersLayer struct {
}

func (p ProxyHeadersLayer) OnRequest(state *LayerState) error {
	p.modifyHeaders(state.RequestHeaders)
	return nil
}

func (p ProxyHeadersLayer) OnResponse(state *LayerState) {
	p.modifyHeaders(state.ResponseHeaders)
}

func (p ProxyHeadersLayer) modifyHeaders(set *HeaderSet) {
	set.DeleteString("proxy-connection")
	set.DeleteString("proxy-authenticate")
	set.DeleteString("proxy-authorization")
	set.DeleteString("connection")
	set.DeleteString("keep-alive")
	set.DeleteString("te")
	set.DeleteString("trailers")
	set.DeleteString("upgrade")
}
