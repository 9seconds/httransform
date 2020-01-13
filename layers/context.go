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

	Tunneled bool

	data map[string]interface{}
	ctx  context.Context
}

func (l *LayerContext) Reset() {
	headers.ReleaseSet(l.RequestHeaders)
	headers.ReleaseSet(l.ResponseHeaders)

	l.Request = nil
	l.RequestHeaders = nil
	l.Response = nil
	l.ResponseHeaders = nil
	l.Tunneled = false
	l.ctx = nil

	for k := range l.data {
		delete(l.data, k)
	}
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

func (l *LayerContext) SetData(key string, value interface{}) {
	l.data[key] = value
}

func (l *LayerContext) GetData(key string) (interface{}, bool) {
	value, ok := l.data[key]

	return value, ok
}

func (l *LayerContext) DeleteData(key string) {
	delete(l.data, key)
}

func (l *LayerContext) SyncResponse() {
	code := l.Response.StatusCode()

	l.syncHeaders(&l.Response.Header, l.ResponseHeaders)
	l.Response.SetStatusCode(code)
}

func (l *LayerContext) SyncRequest() {
	l.syncHeaders(&l.Request.Header, l.RequestHeaders)
}

func (l *LayerContext) ConsumeRequest(req *fasthttp.Request) {
	l.Request = req

	l.RequestHeaders.Reset()
	req.Header.VisitAllInOrder(l.RequestHeaders.SetBytes)
}

func (l *LayerContext) ConsumeResponse(resp *fasthttp.Response) {
	l.Response = resp

	l.ResponseHeaders.Reset()
	resp.Header.VisitAll(l.ResponseHeaders.SetBytes)
}

func (l *LayerContext) syncHeaders(fh fasthttpHeader, set *headers.Set) {
	contentLength := fh.ContentLength()

	fh.Reset()
	fh.DisableNormalizing()

	for _, v := range set.Items() {
		fh.SetBytesKV([]byte(v.Name), []byte(v.Value))
	}

	fh.SetContentLength(contentLength)
}
