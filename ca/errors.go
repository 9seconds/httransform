package ca

import "github.com/9seconds/httransform/v2/errors"

// ErrContextClosed is returned if we ask for the certificate but
// corresponding context was already closed and whole CA should be
// terminated.
var ErrContextClosed = &errors.Error{
	Message: "context is closed",
}
