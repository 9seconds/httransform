package events

import "github.com/valyala/fasthttp"

type RequestMeta struct {
	RequestID string
	Method    string
	URI       fasthttp.URI
}

type ResponseMeta struct {
	Request    *RequestMeta
	StatusCode int
}
