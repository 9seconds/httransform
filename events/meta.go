package events

import (
	"fmt"
	"net"
	"strings"

	"github.com/valyala/fasthttp"
)

type RequestType byte

const (
	RequestTypePlain RequestType = 1 << iota
	RequestTypeTLS
	RequestTypeHTTP
	RequestTypeUpgraded
)

func (r RequestType) String() string {
	var params []string

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

type RequestMeta struct {
	RequestID   string
	Method      string
	URI         fasthttp.URI
	RequestType RequestType
}

func (r *RequestMeta) String() string {
	return fmt.Sprintf("<%s(method=%s uri=%s; params=%v)>",
		r.RequestID,
		r.Method,
		r.URI.String(),
		r.RequestType)
}

type ResponseMeta struct {
	RequestID  string
	StatusCode int
}

func (r *ResponseMeta) String() string {
	return fmt.Sprintf("<%s(status_code=%d)>", r.RequestID, r.StatusCode)
}

type ErrorMeta struct {
	RequestID string
	Err       error
}

func (e *ErrorMeta) String() string {
	return fmt.Sprintf("<%s(error=%v)>", e.RequestID, e.Err)
}

func (e *ErrorMeta) Error() string {
	return e.Err.Error()
}

type TrafficMeta struct {
	ID           string
	Addr         net.Addr
	ReadBytes    uint64
	WrittenBytes uint64
}

func (t *TrafficMeta) String() string {
	return fmt.Sprintf("<%s(addr=%v, read_bytes=%d, written_bytes=%d)>",
		t.ID,
		t.Addr,
		t.ReadBytes,
		t.WrittenBytes)
}
