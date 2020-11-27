package events

import (
	"net"

	"github.com/valyala/fasthttp"
)

type RequestType byte

const (
	RequestTypePlain RequestType = 1 << iota
	RequestTypeTLS
	RequestTypeHTTP
	RequestTypeUpgraded
)

type RequestMeta struct {
	RequestID   string
	Method      string
	URI         fasthttp.URI
	RequestType RequestType
}

type ResponseMeta struct {
	RequestID  string
	StatusCode int
}

type ErrorMeta struct {
	RequestID string
	Error     error
}

type TrafficMeta struct {
	ID           string
	Addr         net.Addr
	ReadBytes    uint64
	WrittenBytes uint64
}
