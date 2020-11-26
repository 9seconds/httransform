package events

import "github.com/valyala/fasthttp"

type RequestMeta struct {
	RequestID string
	Method    string
	URI       fasthttp.URI
}

type ResponseMeta struct {
	RequestID  string
	StatusCode int
}

type ErrorMeta struct {
	RequestID string
	Error     error
}
