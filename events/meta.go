package events

import (
	"fmt"
	"net"
	"strings"

	"github.com/valyala/fasthttp"
)

// RequestType defines a bitmask which corresponds to different
// characteristics of the event.
type RequestType byte

const (
	// RequestTypePlain defines that request is plain HTTP one.
	RequestTypePlain RequestType = 1 << iota

	// RequestTypeTLS defines that request was done via CONNECT + TLS.
	RequestTypeTLS

	// RequestTypeHTTP defines that request is simply HTTP one. Not a
	// websocket or any other upgrade.
	RequestTypeHTTP

	// RequestTypeUpgraded defines that request requires Upgrade
	// of HTTP connection to TCP tunnel. Like websockets.
	RequestTypeUpgraded
)

// String conforms fmt.Stringer interface.
func (r RequestType) String() string {
	params := make([]string, 0, 2)

	if r&RequestTypePlain != 0 {
		params = append(params, "plain")
	}

	if r&RequestTypeTLS != 0 {
		params = append(params, "tls")
	}

	if r&RequestTypeHTTP != 0 {
		params = append(params, "http")
	}

	if r&RequestTypeUpgraded != 0 {
		params = append(params, "upgraded")
	}

	return strings.Join(params, " | ")
}

// RequestMeta defines a metadata related to a request.
type RequestMeta struct {
	// RequestID is unique identifier of the request.
	RequestID string

	// Method is HTTP verb of the request.
	Method string

	// URI is requested URI.
	URI fasthttp.URI

	// User defines a name of the user populated by the given
	// authenticator.
	User string

	// Addr defines an IP address of the user.
	Addr net.Addr

	// RequestType defines a set of characteristics related to that
	// request.
	RequestType RequestType
}

// String conforms fmt.Stringer interface.
func (r *RequestMeta) String() string {
	return fmt.Sprintf("<%s(method=%s uri=%s; params=%v)>",
		r.RequestID,
		r.Method,
		r.URI.String(),
		r.RequestType)
}

// ResponseMeta defines a metadata related to a response.
type ResponseMeta struct {
	// RequestID is unique identifier of the request.
	RequestID string

	// StatusCode is HTTP status code of the response.
	StatusCode int
}

// String conforms fmt.Stringer interface.
func (r *ResponseMeta) String() string {
	return fmt.Sprintf("<%s(status_code=%d)>", r.RequestID, r.StatusCode)
}

// ErrorMeta defines a metadata related to some logical error related to
// the request.
type ErrorMeta struct {
	// RequestID is unique identifier of the request.
	RequestID string

	// Err is an underlying error.
	Err error
}

// Error conforms an error interface.
func (e *ErrorMeta) Error() string {
	return fmt.Sprintf("%s: %v", e.RequestID, e.Err)
}

// Unwrap conforms go1.13 error interface.
func (e *ErrorMeta) Unwrap() error {
	return e.Err
}

// CommonErrorMeta defines a metadata related to some generic
// error produced by HTTP server.
type CommonErrorMeta struct {
	// URI is requested URI.
	URI fasthttp.URI

	// Method defines an HTTP verb.
	Method string

	// Addr defines a remote address of the client.
	Addr net.Addr

	// Err is an underlying error.
	Err error
}

// Error conforms error interface.
func (c *CommonErrorMeta) Error() string {
	return fmt.Sprintf("%s %s (%v): %v", c.URI.String(), c.Method, c.Addr, c.Err)
}

// Unwrap conforms go1.13 error interface.
func (c *CommonErrorMeta) Unwrap() error {
	return c.Err
}

// TrafficMeta defines a metadata related to a TrafficConn completed work.
type TrafficMeta struct {
	// ID a unique identifier of the TrafficConn. Usually it is the same
	// as httransform ones but it is possible to pass your own id here
	// if you use TrafficConn somewhere else.
	ID string

	// Addr defines a 'netloc' IP address which was used.
	Addr net.Addr

	// ReadBytes defines how many bytes were read from the netloc.
	ReadBytes uint64

	// WrittenBytes defines how many bytes were written to the netloc.
	WrittenBytes uint64
}

// String conforms fmt.Stringer interface.
func (t *TrafficMeta) String() string {
	return fmt.Sprintf("<%s(addr=%v, read_bytes=%d, written_bytes=%d)>",
		t.ID,
		t.Addr,
		t.ReadBytes,
		t.WrittenBytes)
}
