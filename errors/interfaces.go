package errors

import "github.com/valyala/fasthttp"

// ErrorJSON is just an interface for error which can render itself into
// JSONs.
type ErrorJSON interface {
	error

	// ErrorJSON returns JSON string of the error.
	ErrorJSON() string
}

// ErrorRenderer is ErrorJSON which an render itself into fasthttp request ctx.
type ErrorRenderer interface {
	error

	WriteTo(*fasthttp.RequestCtx)
}
