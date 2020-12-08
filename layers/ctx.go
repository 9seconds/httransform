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

// RequestHijacker is a function signature you can use for internal
// hijacking. You function will have both ends: a client connection and
// a netloc connection.
type RequestHijacker func(clientConn, netlocConn net.Conn)

// Context is a data structure which we pass along the request. It
// contains requests, responses, different metadata. You can attach some
// free-form data there.
type Context struct {
	// ConnectTo contains an endpoint we have to connect to as a next
	// hop. For example, if you have TLS tunnel, you have 2 addresses:
	// one from CONNECT method, another one - from a tunneled request.
	ConnectTo string

	// RequestID is some unique identifier of the request.
	RequestID string

	// User is a username of a client which is doing a request. It
	// is returned from auth.Interface. If you do not have setup
	// authentication, it is an empty string.
	User string

	// EventStream is an instance of event stream to use.
	EventStream events.Stream

	// RequestType is a bitset related to different characteristics of the
	// request.
	RequestType events.RequestType

	// RequestHeaders is a headers of the request.
	//
	// If you need to do something with headers, please work with this
	// field. Underlying request is case sensitive so it would be really
	// awkward to do anything with it.
	//
	// Underlying fasthttp.Request headers are going to be dropped
	// before sending a request and repopulated from this one.
	RequestHeaders headers.Headers

	// ResponseHeaders is a headers of the response.
	//
	// If you need to do something with headers, please work with this
	// field. Underlying request is case sensitive so it would be really
	// awkward to do anything with it.
	//
	// Underlying fasthttp.Response headers are going to be dropped before
	// sending a response and repopulated from this one.
	ResponseHeaders headers.Headers

	ctxCancel   context.CancelFunc
	ctx         context.Context
	originalCtx *fasthttp.RequestCtx
	values      map[string]interface{}
}

// Request returns a pointer to the original fasthttp.Request.
func (c *Context) Request() *fasthttp.Request {
	return &c.originalCtx.Request
}

// Response returns a pointer to the original fasthttp.Response.
func (c *Context) Response() *fasthttp.Response {
	return &c.originalCtx.Response
}

// RemoteAddr returns an instance of remote address of the client.
func (c *Context) RemoteAddr() net.Addr {
	return c.originalCtx.RemoteAddr()
}

// LocalAddr returns an instance of local address of the client.
func (c *Context) LocalAddr() net.Addr {
	return c.originalCtx.LocalAddr()
}

// Respond is just a shortcut for the fast response. This response is
// just a plain text with a status code.
func (c *Context) Respond(msg string, statusCode int) {
	c.originalCtx.Response.Reset()
	c.originalCtx.Response.Header.DisableNormalizing()
	c.originalCtx.Response.SetStatusCode(statusCode)
	c.originalCtx.Response.SetConnectionClose()
	c.originalCtx.Response.Header.SetContentType("text/plain")
	c.originalCtx.Response.SetBodyString(msg)
}

// Error is a shortcut for the fast response about given error.
func (c *Context) Error(err error) {
	c.Respond(fmt.Sprintf("Request has failed: %s", err.Error()),
		fasthttp.StatusInternalServerError)
}

// Hijack setups a hijacker for the request if necessary. Usually
// you need it if you want to process upgraded connections such as
// websockets.
//
// netlocConn is a connection to a target website. It is closed when you
// exit a hijacker function.
func (c *Context) Hijack(netlocConn net.Conn, hijacker RequestHijacker) {
	handler := conns.FixHijackHandler(func(clientConn net.Conn) bool {
		defer netlocConn.Close()

		hijacker(clientConn, netlocConn)

		return true
	})
	c.originalCtx.Hijack(handler)
}

// Hijacked checks if given context was hijacked or not.
func (c *Context) Hijacked() bool {
	return c.originalCtx.Hijacked()
}

// Init initializes a Context based on given parameters.
//
// fasthttpCtx is a parent context which produced this one, connectTo
// is a host:port of the remote endpoint we need to connect to on a
// first-hop (think about CONNECT method). The rest of parameters are
// trivial.
func (c *Context) Init(fasthttpCtx *fasthttp.RequestCtx,
	connectTo string,
	eventStream events.Stream,
	user string,
	requestType events.RequestType) {
	ctx, cancel := context.WithCancel(fasthttpCtx)

	c.RequestID = uuid.Must(uuid.NewV4()).String()
	c.RequestType = requestType
	c.EventStream = eventStream
	c.ConnectTo = connectTo
	c.User = user

	c.originalCtx = fasthttpCtx
	c.ctx = ctx
	c.ctxCancel = cancel

	uri := fasthttpCtx.Request.URI()

	if requestType.IsTLS() {
		uri.SetScheme("https")
	} else {
		uri.SetScheme("http")
	}
}

// Reset resets a state of the given context. It also cancels it if
// necessary.
func (c Context) Reset() {
    if c.ctxCancel != nil {
        c.ctxCancel()
    }

	c.originalCtx = nil
	c.ctx = nil
	c.ctxCancel = nil

	c.RequestID = ""
	c.RequestType = 0
	c.EventStream = nil
	c.ConnectTo = ""
	c.User = ""

	c.RequestHeaders.Reset()
	c.ResponseHeaders.Reset()

	for key := range c.values {
		delete(c.values, key)
	}
}

// Cancel cancels a given context.
func (c *Context) Cancel() {
	c.ctxCancel()
}

// Deadline conforms a context.Context interface.
func (c *Context) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

// Done conforms a context.Context interface.
func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err conforms a context.Context interface.
func (c *Context) Err() error {
	return c.ctx.Err()
}

// Value conforms a context.Context interface.
func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// Set sets a value to internal storage.
func (c *Context) Set(key string, value interface{}) {
	c.values[key] = value
}

// Get gets a value from internal storage
func (c *Context) Get(key string) (interface{}, bool) {
	value, ok := c.values[key]

	return value, ok
}

// Delete deletes a key from the internal storage.
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

// AcquireContext returns a new context from the pool.
func AcquireContext() *Context {
	return poolContext.Get().(*Context)
}

// ReleaseContext returns a context back to the pool.
func ReleaseContext(ctx *Context) {
	ctx.Reset()
	poolContext.Put(ctx)
}
