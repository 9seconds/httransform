package events

import (
	"fmt"
	"net"

	"github.com/valyala/fasthttp"
)

// RequestType defines a bitmask which corresponds to different
// characteristics of the event.
type RequestType byte

const (
	RequestTypeTunneled RequestType = 1 << iota
	RequestTypeTLS
	RequestTypeUpgraded
)

func (r RequestType) IsTunneled() bool {
	return r&RequestTypeTunneled != 0
}

func (r RequestType) IsTLS() bool {
	return r&RequestTypeTLS != 0
}

func (r RequestType) IsUpgraded() bool {
	return r&RequestTypeUpgraded != 0
}

// String conforms fmt.Stringer interface.
func (r RequestType) String() string {
	tunneled := "tunneled:no"
	tls := "tls:no"
	upgraded := "upgraded:no"

	if r.IsTunneled() {
		tunneled = "tunneled:yes"
	}

	if r.IsTLS() {
		tls = "tls:yes"
	}

	if r.IsUpgraded() {
		upgraded = "upgraded:yes"
	}

	return tunneled + " | " + tls + " | " + upgraded
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
	return fmt.Sprintf("<%s(method=%s, user=%s, remote_addr=%v, uri=%s, params=%v)>",
		r.RequestID,
		r.Method,
		r.User,
		r.Addr,
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
	return e.RequestID + ": " + e.Err.Error()
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
