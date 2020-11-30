package auth

import "github.com/valyala/fasthttp"

// NoopAuth is a dummy implementation which always welcomes everyone.
// A username is empty though.
type NoopAuth struct{}

// Authenticate does auth based on given context.
func (n NoopAuth) Authenticate(_ *fasthttp.RequestCtx) (string, error) {
	return "", nil
}
