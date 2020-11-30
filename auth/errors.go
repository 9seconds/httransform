package auth

import "errors"

var (
	ErrMalformedHeaderValue = errors.New("header value is malformed")
	ErrFailedAuth           = errors.New("authorization failed")
	ErrAuthRequired         = errors.New("authentication is required")
)
