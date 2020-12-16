package auth

import "github.com/valyala/fasthttp"

// Interface defines a set of methods which are mandatory for each
// authenicator to implement.
//
// Authenticator goal is to take a 'raw' fasthttp's context and
// return a pair of username and error. If error is not nil, it means
// authentication has failed. If nil, then everything is fine and
// username value contains a name (or identifier of the user to use
// next).
//
// All implementation SHOULD work with RequestCtx where normalization is
// disabled.
type Interface interface {
	// Authenticate does auth based on given context.
	Authenticate(*fasthttp.RequestCtx) (string, error)
}
