package auth

import "errors"

var (
	// ErrMalformedHeaderValue is returned if header is malformed. So,
	// header value is present but format is incorrect: no 'basic ' prefix
	// which defines a basic auth schema and so on.
	ErrMalformedHeaderValue = errors.New("header value is malformed")

	// ErrFailedAuth is returned if authorization is failed. It means that
	// we were able to reach it but a core logic has failed.
	ErrFailedAuth = errors.New("authorization failed")

	// ErrAuthRequired means that user gave no credentials.
	ErrAuthRequired = errors.New("authentication is required")
)
