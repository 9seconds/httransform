package httransform

import (
	"crypto/subtle"

	"github.com/valyala/fasthttp"
)

// Layer is the middleware for HTTP proxy. Every request will go through
// OnRequest method. If method spot the error, it has to return it. In
// that case we won't propagate such error forward, it is going to go
// backwards by the stack of layers.
//
// We do not limit you but encourgate every layer to manage its own
// errors. So, if your layer raises ErrSomething, then you have to
// handle only ErrSomething in OnResponse method. Any other errors
// should be ignored. Sometimes this advice is tricky to implement but
// in general it leads to more simpler design.
//
// Layer contracts: if you executed OnRequest, then we should execute
// OnResponse. Another contract is order: if you execute layer in order
// 1-2-3-4-5, then OnResponse chains will be executed in order of
// 5-4-3-2-1 (stack rewind). If step 3 executed error, then execution is
// stopped and we'll call OnResponses as 3-2-1.
type Layer interface {
	OnRequest(*LayerState) error
	OnResponse(*LayerState, error)
}

// LayerState is the state structure which propagated through layer
// chain and terminated on Executor.
//
// You may see Request/Response structures there. You can update then
// but we encourage you to update relevant HeaderSet structure if you
// want to keep order of headers.
//
// Also, LayerState provides context-like interface so you can store
// your values there using Set/Get/Delete methods.
type LayerState struct {
	// RequestID has unique identifier of every request.
	RequestID uint64

	// ProxyUser is username taken from Proxy-Authorization header. Empty
	// if nothing is present.
	ProxyUser []byte

	// ProxyPassword is password taken from Proxy-Authorization header.
	// Empty if nothing is present.
	ProxyPassword []byte

	// RequestHeaders contains request headers. This structure is populated
	// by httransform so you do not need to fill it from the inital
	// contents of Request.
	//
	// If you want to update request headers, you should do it using
	// this field, do not modify Request directly. After the last layer
	// but before propagation to executer, httransform drops contents of
	// Request.Header and fills them with items from RequestHeaders.
	RequestHeaders *HeaderSet

	// ResponseHeaders contains response headers. This structure is
	// populated by httransform so you do not need to fill it from the
	// inital contents of Response.
	//
	// If you want to update response headers, you should do it using this
	// field, do not modify Response directly. After the last layer but
	// before returning back to the client, httransform drops contents of
	// Response.Header and fills them with items from ResponseHeaders.
	ResponseHeaders *HeaderSet

	// Request is the instance of the original request.
	Request *fasthttp.Request

	// Response is the instance of the original response.
	Response *fasthttp.Response

	isConnect bool
	ctx       map[string]interface{}
}

// Set sets new value to the context of layer state.
func (l *LayerState) Set(key string, value interface{}) {
	l.ctx[key] = value
}

// Get takes value from the context of layer state. If LayerState
// instance has no such key, then last parameter is false. If it has -
// true.
func (l *LayerState) Get(key string) (interface{}, bool) {
	item, ok := l.ctx[key]
	return item, ok
}

// Delete removes key from the context of the LayerState. It is safe
// to unknown keys.
func (l *LayerState) Delete(key string) {
	delete(l.ctx, key)
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

// AddRemoveHeaderLayer is the generic layer to manage request/response
// headers. If you need to add some special headers or remove sensitive
// data, please use this layer.
type AddRemoveHeaderLayer struct {
	// FrozenRequestHeaders is the set of headers which should be present
	// in the request with given values.
	FrozenRequestHeaders map[string]string

	// FrozenResponseHeaders is the set of headers which should be present
	// in the response with given values.
	FrozenResponseHeaders map[string]string

	// DefaultRequestHeaders is the set of headers which should be present
	// in the request. If RequestHeaders already has some value for the
	// header, then this layer does nothing. If value is absent, then
	// header is populated with the given default.
	DefaultRequestHeaders map[string]string

	// DefaultResponseHeaders is the set of headers which should be present
	// in the response. If ResponseHeaders already has some value for the
	// header, then this layer does nothing. If value is absent, then
	// header is populated with the given default.
	DefaultResponseHeaders map[string]string

	// AbsentRequestHeaders is the list of headers which should be
	// dropped from the request. This list has bigger priority than
	// FrozenRequestHeaders or DefaultRequestHeaders.
	AbsentRequestHeaders []string

	// AbsentResponseHeaders is the list of headers which should be
	// dropped from the response. This list has bigger priority than
	// FrozenResponseHeaders or DefaultResponseHeaders.
	AbsentResponseHeaders []string
}

// OnRequest is the callback for Layer interface.
func (a *AddRemoveHeaderLayer) OnRequest(state *LayerState) error {
	a.apply(state.RequestHeaders, a.FrozenRequestHeaders, a.DefaultRequestHeaders, a.AbsentRequestHeaders)
	return nil
}

// OnResponse is the callback for Layer interface.
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
