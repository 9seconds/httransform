package ca

import "errors"

// ErrContextClosed is returned if we ask for the certificate but
// corresponding context was already closed and whole CA should be
// terminated.
var ErrContextClosed = errors.New("context is closed")
