package layers

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type Context struct {
	requestID   string
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

func (c *Context) RequestID() string {
	return c.requestID
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
	c.Respond(fmt.Sprintf("Request has failed: %s", err.Error()), fasthttp.StatusInternalServerError)
}

func (c *Context) Init(fasthttpCtx *fasthttp.RequestCtx, tunneled bool) {
	ctx, cancel := context.WithCancel(fasthttpCtx)

	c.originalCtx = fasthttpCtx
	c.ctx = ctx
	c.ctxCancel = cancel
	c.requestID = strings.ToLower(strconv.FormatUint(fasthttpCtx.ID(), 16))

	uri := fasthttpCtx.Request.URI()

	if tunneled {
		uri.SetScheme("https")
	} else {
		uri.SetScheme("http")
	}
}

func (c Context) Reset() {
	c.originalCtx = nil
	c.ctx = nil
	c.ctxCancel = nil
	c.requestID = ""

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
