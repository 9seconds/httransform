package events

import "github.com/valyala/fasthttp"

type RequestMeta struct {
	RequestID string
	Method    string
	URI       fasthttp.URI
}

func (r *RequestMeta) ID() string {
	return r.RequestID
}

type ResponseMeta struct {
	Request    *RequestMeta
	StatusCode int
}

func (r *ResponseMeta) ID() string {
	return r.Request.RequestID
}

type ErrorMeta struct {
	Request *RequestMeta
	Error   error
}

func (e *ErrorMeta) ID() string {
	return e.Request.RequestID
}
