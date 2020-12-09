package auth

import (
	"github.com/9seconds/httransform/v2/errors"
	"github.com/valyala/fasthttp"
)

var (
	// ErrFailedAuth is returned if authorization is failed. It means that
	// we were able to reach it but a core logic has failed.
	ErrFailedAuth = &errors.Error{
		Message:    "authorization failed",
		Code:       "bad_auth",
		StatusCode: fasthttp.StatusProxyAuthRequired,
	}

	// ErrAuthRequired means that user gave no credentials.
	ErrAuthRequired = &errors.Error{
		Message:    "authentication is required",
		Code:       "auth_required",
		StatusCode: fasthttp.StatusProxyAuthRequired,
	}
)
