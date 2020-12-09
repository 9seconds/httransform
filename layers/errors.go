package layers

import (
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"
)

const (
	ErrorCodeGenericError   = "generic_error"
	ErrorCodeSubnetFiltered = "subnets_filtered"
	ErrorCodeExecutor       = "executor"
)

// Error defines a custom error which can be returned from a layer. If
// layer returns this error, the rest of machinery can use it to specify
// some status codes or content types.
type Error struct {
	// StatusCode defines a status code which should be returned to a
	// client.
	StatusCode int

	// Details defines some additional details for the error
	Details string

	// An error code of the error to keep in logs or return in headers.
	ErrorCode string

	// Err keeps an original error.
	Err error
}

func (e *Error) AsJSON() string {
	originalError := ""
	if e != nil && e.Err != nil {
		originalError = e.Err.Error()
	}

	rv := struct {
		ErrorCode     string   `json:"error_code"`
		AllErrorCodes []string `json:"all_error_codes"`
		Details       string   `json:"details"`
		Error         string   `json:"error"`
	}{
		ErrorCode:     e.GetErrorCode(),
		AllErrorCodes: e.GetErrors(),
		Details:       e.GetDetails,
		Error:         originalError,
	}

	data, _ := json.Marshal(rv)

	return string(data)
}

// Error to conform error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	message := ""

	if e.Err != nil {
		message += e.Err.Error() + ": "
	}

	return message + e.Details
}

// Unwrap to conform go1.13 error interface.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// GetStatusCode returns a status code for the error.
func (e *Error) GetStatusCode() int {
	if e == nil || e.StatusCode == 0 {
		return fasthttp.StatusInternalServerError
	}

	return e.StatusCode
}

func (e *Error) GetDetails() string {
	if e == nil {
		return ""
	}

	return e.Details
}

func (e *Error) GetErrorCode() string {
	if e == nil {
		return ""
	}

	return e.ErrorCode
}

func (e *Error) GetErrors() []string {
    if e == nil {
        return nil
    }

	errs := []string{}

	var err error = e

	for err != nil {
		if ee, ok := err.(*Error); ok && ee.ErrorCode != "" { // nolint: errorlint
			errs = append(errs, ee.ErrorCode)
		}

		err = errors.Unwrap(err)
	}

	return errs
}

func WrapError(errorCode string, err error) *Error {
	if err == nil {
		return nil
	}

	var customErr *Error

	newErr := &Error{
		ErrorCode: errorCode,
		Err:       err,
	}

	if errors.As(err, &customErr) {
		newErr.StatusCode = customErr.StatusCode
	}

	return newErr
}
