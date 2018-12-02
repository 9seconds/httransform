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

type AddRemoveHeaderLayer struct {
	FrozenRequestHeaders   map[string]string
	FrozenResponseHeaders  map[string]string
	DefaultRequestHeaders  map[string]string
	DefaultResponseHeaders map[string]string
	AbsentRequestHeaders   []string
	AbsentResponseHeaders  []string
}

func (a *AddRemoveHeaderLayer) OnRequest(state *LayerState) error {
	a.apply(state.RequestHeaders, a.FrozenRequestHeaders, a.DefaultRequestHeaders, a.AbsentRequestHeaders)
	return nil
}

func (a *AddRemoveHeaderLayer) OnResponse(state *LayerState, err error) {
	a.apply(state.ResponseHeaders, a.FrozenResponseHeaders, a.DefaultResponseHeaders, a.AbsentResponseHeaders)
}

func (a *AddRemoveHeaderLayer) apply(set *HeaderSet, frozen, defaults map[string]string, absent []string) {
	for key, value := range defaults {
		if _, ok := set.GetString(key); !ok {
			set.SetString(key, value)
		}
	}
	for key, value := range frozen {
		set.SetString(key, value)
	}
	for _, key := range absent {
		set.DeleteString(key)
	}
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
