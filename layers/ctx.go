package layers

import (
	"context"
	"net"
	"strconv"
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
	tunneled    bool
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

func (c *Context) Tunneled() bool {
	return c.tunneled
}

func (c *Context) RequestID() string {
	return c.requestID
}

func (c *Context) Init(fasthttpCtx *fasthttp.RequestCtx, parent *Context) {
	ctx, cancel := context.WithCancel(fasthttpCtx)

	c.originalCtx = fasthttpCtx
	c.ctx = ctx
	c.ctxCancel = cancel

	if parent != nil {
		c.requestID = parent.requestID
		c.tunneled = true
	} else {
		c.requestID = strconv.FormatUint(fasthttpCtx.ID(), 16)
		c.tunneled = false
	}
}

func (c Context) Reset() {
	c.originalCtx = nil
	c.ctx = nil
	c.ctxCancel = nil
	c.requestID = ""
	c.tunneled = false

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
