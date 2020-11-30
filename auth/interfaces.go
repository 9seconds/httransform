package auth

import "github.com/valyala/fasthttp"

type Interface interface {
	Authenticate(*fasthttp.RequestCtx) (string, error)
}
