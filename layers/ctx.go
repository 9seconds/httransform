package layers

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/9seconds/httransform/v2/conns"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/headers"
	"github.com/gofrs/uuid"
	"github.com/valyala/fasthttp"
)

type RequestHijacker func(clientConn, netlocConn net.Conn)

type Context struct {
	ConnectTo       string
	RequestID       string
	Events          events.EventChannel
	RequestHeaders  headers.Headers
	ResponseHeaders headers.Headers

	ctxCancel   context.CancelFunc
	ctx         context.Context
	originalCtx *fasthttp.RequestCtx
	values      map[string]interface{}
}

func (c *Context) Request() *fasthttp.Request {
	return &c.originalCtx.Request
}

func (c *Context) Response() *fasthttp.Response {
	return &c.originalCtx.Response
}

func (c *Context) RemoteAddr() net.Addr {
	return c.originalCtx.RemoteAddr()
}

func (c *Context) LocalAddr() net.Addr {
	return c.originalCtx.LocalAddr()
}

func (c *Context) Respond(msg string, statusCode int) {
	c.originalCtx.Response.Reset()
	c.originalCtx.Response.Header.DisableNormalizing()
	c.originalCtx.Response.SetStatusCode(statusCode)
	c.originalCtx.Response.SetConnectionClose()
	c.originalCtx.Response.Header.SetContentType("text/plain")
	c.originalCtx.Response.SetBodyString(msg)
}

func (c *Context) Error(err error) {
	c.Respond(fmt.Sprintf("Request has failed: %s", err.Error()),
		fasthttp.StatusInternalServerError)
}

func (c *Context) Hijack(netlocConn net.Conn, hijacker RequestHijacker) {
	handler := conns.FixedHijackHandler(func(clientConn net.Conn) bool {
		defer netlocConn.Close()

		hijacker(clientConn, netlocConn)

		return true
	})
	c.originalCtx.Hijack(handler)
}

func (c *Context) Hijacked() bool {
	return c.originalCtx.Hijacked()
}

func (c *Context) Init(fasthttpCtx *fasthttp.RequestCtx,
	connectTo string,
	channelEvents events.EventChannel,
	isTLS bool) {
	ctx, cancel := context.WithCancel(fasthttpCtx)

	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	c.RequestID = id.String()
	c.Events = channelEvents
	c.ConnectTo = connectTo

	c.originalCtx = fasthttpCtx
	c.ctx = ctx
	c.ctxCancel = cancel

	uri := fasthttpCtx.Request.URI()

	if isTLS {
		uri.SetScheme("https")
	} else {
		uri.SetScheme("http")
	}
}

func (c Context) Reset() {
	c.originalCtx = nil
	c.ctx = nil
	c.ctxCancel = nil

	c.RequestID = ""
	c.Events = nil
	c.ConnectTo = ""

	c.RequestHeaders.Reset()
	c.ResponseHeaders.Reset()

	for key := range c.values {
		delete(c.values, key)
	}
}

func (c *Context) Cancel() {
	c.ctxCancel()
}

func (c *Context) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Context) Err() error {
	return c.ctx.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *Context) Set(key string, value interface{}) {
	c.values[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	value, ok := c.values[key]

	return value, ok
}

func (c *Context) Delete(key string) {
	delete(c.values, key)
}

var poolContext = sync.Pool{
	New: func() interface{} {
		return &Context{
			RequestHeaders: headers.Headers{
				Headers: []headers.Header{},
			},
			ResponseHeaders: headers.Headers{
				Headers: []headers.Header{},
			},
			values: map[string]interface{}{},
		}
	},
}

func AcquireContext() *Context {
	return poolContext.Get().(*Context)
}

func ReleaseContext(ctx *Context) {
	ctx.Reset()
	poolContext.Put(ctx)
}
