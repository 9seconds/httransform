package main

import (
	"bytes"
	"encoding/base64"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

type Layer interface {
	OnRequest(*LayerState) error
	OnResponse(*LayerState)
}

type LayerState struct {
	isConnect bool
	ctx       map[string]interface{}

	Error           error
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
	state.Error = nil
	state.RequestID = ctx.ID()
	state.RequestHeaders = requestHeaders
	state.ResponseHeaders = responseHeaders
	state.Request = &ctx.Request
	state.Response = &ctx.Response
	state.isConnect = isConnect
}

type ConnectionCloseLayer struct {
}

func (c *ConnectionCloseLayer) OnRequest(state *LayerState) error {
	return nil
}

func (c *ConnectionCloseLayer) OnResponse(state *LayerState) {
	state.ResponseHeaders.SetString("Connection", "close")
}

type ProxyHeadersLayer struct {
}

func (p *ProxyHeadersLayer) OnRequest(state *LayerState) error {
	p.modifyHeaders(state.RequestHeaders)
	return nil
}

func (p *ProxyHeadersLayer) OnResponse(state *LayerState) {
	p.modifyHeaders(state.ResponseHeaders)
}

func (p *ProxyHeadersLayer) modifyHeaders(set *HeaderSet) {
	set.DeleteString("proxy-connection")
	set.DeleteString("proxy-authenticate")
	set.DeleteString("proxy-authorization")
	set.DeleteString("connection")
	set.DeleteString("keep-alive")
	set.DeleteString("te")
	set.DeleteString("trailers")
	set.DeleteString("upgrade")
}

type ProxyAuthorizationBasicLayer struct {
	Mandatory bool
}

func (p *ProxyAuthorizationBasicLayer) OnRequest(state *LayerState) error {
	username, password, err := p.extract(state)
	if err == nil {
		state.Set("username", username)
		state.Set("password", password)
		return nil
	}

	if p.Mandatory {
		return ProxyAuthorizationError
	}

	return nil
}

func (p *ProxyAuthorizationBasicLayer) OnResponse(state *LayerState) {
	if state.Error == ProxyAuthorizationError {
		MakeBadResponse(state.Response, "", fasthttp.StatusProxyAuthRequired)
	}
}

func (p *ProxyAuthorizationBasicLayer) extract(state *LayerState) (string, string, error) {
	line, ok := state.RequestHeaders.GetBytes([]byte("proxy-authorization"))
	if !ok {
		return "", "", errors.New("Cannot extract contents of Proxy-Authorization header")
	}

	pos := bytes.IndexByte(line, ' ')
	if pos < 0 {
		return "", "", errors.New("Malformed Proxy-Authorization header")
	}
	if !bytes.HasPrefix(bytes.ToLower(line[:pos]), []byte("basic")) {
		return "", "", errors.New("Incorrect authorization prefix")
	}

	for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t') {
		pos++
	}

	line, err := base64.StdEncoding.DecodeString(string(line[pos:]))
	if err != nil {
		return "", "", errors.Annotate(err, "Incorrectly encoded authorization payload")
	}

	pos = bytes.IndexByte(line, ':')
	if pos < 0 {
		return "", "", errors.New("Cannot find a user/password delimiter in decoded authorization string")
	}

	return string(line[:pos]), string(line[pos+1:]), nil
}
