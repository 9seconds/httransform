package layers

import (
	"context"
	"net"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/headers"
)

type LayerContext struct {
	RequestID uint64

	RequestHeaders  *headers.Set
	ResponseHeaders *headers.Set

	Request  *fasthttp.Request
	Response *fasthttp.Response

	LocalAddr  net.Addr
	RemoteAddr net.Addr

	data map[string]interface{}
	ctx  context.Context
}

func (l *LayerContext) Reset() {
	headers.ReleaseSet(l.RequestHeaders)
	headers.ReleaseSet(l.ResponseHeaders)

	l.RequestHeaders = nil
	l.ResponseHeaders = nil
	l.Request = nil
	l.Response = nil
	l.LocalAddr = nil
	l.RemoteAddr = nil

	for k := range l.data {
		delete(l.data, k)
	}

	l.ctx = nil
}

func (l *LayerContext) Deadline() (time.Time, bool) {
	return l.ctx.Deadline()
}

func (l *LayerContext) Done() <-chan struct{} {
	return l.ctx.Done()
}

func (l *LayerContext) Err() error {
	return l.ctx.Err()
}

func (l *LayerContext) Value(key interface{}) interface{} {
	return l.ctx.Value(key)
}

func (l *LayerContext) Set(key string, value interface{}) {
	l.data[key] = value
}

func (l *LayerContext) Get(key string) (interface{}, bool) {
	value, ok := l.data[key]

	return value, ok
}

func (l *LayerContext) Delete(key string) {
	delete(l.data, key)
}

func (l *LayerContext) ConsumeRequest(request *fasthttp.Request) {
	l.RequestHeaders.Reset()
	request.Header.VisitAllInOrder(func(key, value []byte) {
		l.RequestHeaders.Add(string(key), string(value))
	})
}

func (l *LayerContext) ConsumeResponse(response *fasthttp.Response) {
	l.ResponseHeaders.Reset()
	response.Header.VisitAll(func(key, value []byte) {
		l.ResponseHeaders.Add(string(key), string(value))
	})
}

func (l *LayerContext) SyncToRequest() {
	syncHeaders(&l.Request.Header, l.RequestHeaders)
}

func (l *LayerContext) SyncToResponse() {
	statusCode := l.Response.StatusCode()

	syncHeaders(&l.Response.Header, l.ResponseHeaders)
	l.Response.SetStatusCode(statusCode)
}

func (l *LayerContext) SetSimpleResponse(statusCode int, message string) {
	l.Response.Reset()
	l.Response.SetBodyString(message)
	l.Response.SetStatusCode(statusCode)
	l.Response.Header.SetContentType("text/plain")
}

func syncHeaders(fh fasthttpHeader, set *headers.Set) {
	contentLength := fh.ContentLength()
	fh.Reset()
	fh.DisableNormalizing()

	for _, v := range set.All() {
		fh.Add(v.Name, v.Value)
	}

	fh.SetContentLength(contentLength)
}
