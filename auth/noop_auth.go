package auth

import "github.com/valyala/fasthttp"

type NoopAuth struct{}

func (n NoopAuth) Authenticate(_ *fasthttp.RequestCtx) (string, error) {
	return "", nil
}
