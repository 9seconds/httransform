package errors

import stderrors "errors"

// New created a new Error from this package.
func New(text string) error {
	return &Error{
		Message: text,
	}
}

// Unwrap is just calling stderrors.Unwrap. It is here for
// convenience.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// As is just calling stderrors.As. It is here for convenience.
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

// Is is just calling stderrors.Is. It is here for convenience.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// Annotate wraps/chains error into new Error instance.
func Annotate(err error, message, code string, statusCode int) *Error {
	return &Error{
		Code:       code,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
