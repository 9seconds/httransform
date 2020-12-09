package errors

import "github.com/valyala/fasthttp"

type ErrorJSON interface {
	error

	ErrorJSON() string
}

type ErrorJSONRenderer interface {
	ErrorJSON

	WriteTo(*fasthttp.RequestCtx)
}
