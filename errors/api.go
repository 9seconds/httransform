package errors

import stderrors "errors"

func New(text string) error {
	return &Error{
		Message: text,
	}
}

func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

func Annotate(err error, message, code string, statusCode int) *Error {
	return &Error{
		Code:       code,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
