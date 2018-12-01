package httransform

import (
	"crypto/subtle"

	"github.com/valyala/fasthttp"
)

type Layer interface {
	OnRequest(*LayerState) error
	OnResponse(*LayerState, error)
}

type LayerState struct {
	isConnect bool
	ctx       map[string]interface{}

	RequestID       uint64
	ProxyUser       []byte
	ProxyPassword   []byte
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

func initLayerState(state *LayerState, ctx *fasthttp.RequestCtx,
	requestHeaders, responseHeaders *HeaderSet,
	isConnect bool, user, password []byte) {
	state.RequestID = ctx.ID()
	state.RequestHeaders = requestHeaders
	state.ResponseHeaders = responseHeaders
	state.Request = &ctx.Request
	state.Response = &ctx.Response
	state.isConnect = isConnect
	state.ProxyUser = user
	state.ProxyPassword = password
}

type ConnectionCloseLayer struct {
}

func (c *ConnectionCloseLayer) OnRequest(_ *LayerState) error {
	return nil
}

func (c *ConnectionCloseLayer) OnResponse(state *LayerState, _ error) {
	state.ResponseHeaders.SetString("Connection", "close")
	state.ResponseHeaders.SetString("Proxy-Connection", "close")
}

type ProxyHeadersLayer struct {
}

func (p *ProxyHeadersLayer) OnRequest(state *LayerState) error {
	p.modifyHeaders(state.RequestHeaders)
	return nil
}

func (p *ProxyHeadersLayer) OnResponse(state *LayerState, _ error) {
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
}

type ProxyAuthorizationBasicLayer struct {
	User     []byte
	Password []byte
	Realm    string
}

func (p *ProxyAuthorizationBasicLayer) OnRequest(state *LayerState) error {
	userFault := subtle.ConstantTimeCompare(state.ProxyUser, p.User) != 1
	passwordFault := subtle.ConstantTimeCompare(state.ProxyPassword, p.Password) != 1
	if userFault || passwordFault {
		return ErrProxyAuthorization
	}

	return nil
}

func (p *ProxyAuthorizationBasicLayer) OnResponse(state *LayerState, err error) {
	if err == ErrProxyAuthorization {
		p.MakeProxyAuthRequiredResponse(state)
	}
}

func (p *ProxyAuthorizationBasicLayer) MakeProxyAuthRequiredResponse(state *LayerState) {
	MakeSimpleResponse(state.Response, "", fasthttp.StatusProxyAuthRequired)

	var authString string
	if p.Realm != "" {
		authString = "Basic realm=\"" + p.Realm + "\""
	} else {
		authString = "Basic"
	}
	state.ResponseHeaders.SetString("Proxy-Authenticate", authString)
}
