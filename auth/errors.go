package auth

import "errors"

var (
	// ErrFailedAuth is returned if authorization is failed. It means that
	// we were able to reach it but a core logic has failed.
	ErrFailedAuth = errors.New("authorization failed")

	// ErrAuthRequired means that user gave no credentials.
	ErrAuthRequired = errors.New("authentication is required")
)
