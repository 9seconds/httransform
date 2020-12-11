package errors

import (
	"strings"

	"github.com/valyala/fasthttp"
)

// Error defines a custom error which can be returned from a layer or
// executor. This error has a stack of attached errors and can render
// JSON. Also, it keeps a status code so if you are searching for a
// correct way of setting your own response code, this is a best one.
//
// It is important to keep Error chain consistent. If we have an error
// chain like Error -> Error -> ??? -> Error -> ??? -> Error, then only
// first 2 errors are going to be used for JSON generation.
type Error struct {
	// StatusCode defines a status code which should be returned to a
	// client.
	StatusCode int

	// Message defines some custom message which is specific for this
	// error.
	Message string

	// Code defines a special code for machine-readable processing.
	Code string

	// Err keeps an original error
	Err error
}

// Error to conform error interface.
func (e *Error) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Err != nil && e.Message != "":
		return e.Message + ": " + e.Err.Error()
	case e.Err != nil:
		return e.Err.Error()
	}

	return e.Message
}

// Unwrap to conform go 1.13 error interface.
func (e *Error) Unwrap() error {
	if e != nil {
		return e.Err
	}

	return nil
}

// GetStatusCode returns a status code of THIS error (not a chain one).
// It also works with nil pointers so it is a preferrable way of getting
// a data out of this error.
func (e *Error) GetStatusCode() int {
	if e != nil {
		return e.StatusCode
	}

	return 0
}

// GetChainStatusCode returns a status code of the whole error chain.
//
// Lets assume that we have errors A, B. A wraps B, B wraps C. Status
// code of A is 0 (default one), B - 400, C - 500. So, we have a chain
// of statuses 0 -> 400 -> 500. GetChainStatusCode returns a first != 0
// status (400 in our case). If whole chain is empty, it returns 500.
func (e *Error) GetChainStatusCode() int {
	for current := e; current != nil; current = unwrapError(current) {
		if current.StatusCode != 0 {
			return current.StatusCode
		}
	}

	return fasthttp.StatusInternalServerError
}

// GetMessage returns a message of THIS error (not a chain one). It also
// works with nil pointers so it is a preferrable way of getting a data
// out of this error.
func (e *Error) GetMessage() string {
	if e != nil {
		return e.Message
	}

	return ""
}

// GetCode returns a code of THIS error (not a chain one). It also works
// with nil pointers so it is a preferrable way of getting a data out of
// this error.
func (e *Error) GetCode() string {
	if e != nil {
		return e.Code
	}

	return ""
}

// GetChainCode returns an error code of the whole error chain.
//
// Lets assume that we have errors A, B. A wraps B, B wraps C. Code of
// A is "" (default one), B - "foo", C - "bar". So, we have a chain
// of statuses "" -> "foo" -> "bar". GetChainCode returns a first !=
// "" status ("foo" in our case). If whole chain is empty, it returns
// "internal_error".
func (e *Error) GetChainCode() string {
	for current := e; current != nil; current = unwrapError(current) {
		if e.Code != "" {
			return e.Code
		}
	}

	return "internal_error"
}

// ErrorJSON returns a JSON encoded representation of the error.
//
// JSON has a following structure:
//
//   {
//     "error": {
//       "code": "executor",
//       "message": "cannot execute a request",
//       "stack": [
//         {
//           "code": "executor",
//           "message": "cannot execute a request",
//           "status_code": 0
//         },
//         {
//           "code": "",
//           "message": "cannot dial to the netloc",
//           "status_code": 0
//         },
//         {
//           "code": "",
//           "message": "cannot upgrade connection to tls",
//           "status_code": 0
//         },
//         {
//           "code": "tls_handshake",
//           "message": "cannot perform TLS handshake",
//           "status_code": 0
//         },
//         {
//           "code": "",
//           "message": "x509: cannot validate certificate for 23.23.154.131 because it doesn't contain any IP SANs",
//           "status_code": 0
//         }
//       ]
//     }
//   }
//
// * code is a first chain code of the error
//
// * message is own error message
//
// * stack has a list of errors (last error is initiator) with similar structure.
func (e *Error) ErrorJSON() string {
	builder := strings.Builder{}

	writeErrorAsJSON(e, &builder)

	value := builder.String()
    value = value[:len(value)-1]

	return value
}

// WriteTo writes this error into a given fasthttp request context.
func (e *Error) WriteTo(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.Response.Header.DisableNormalizing()
	ctx.Response.SetConnectionClose()
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.Header.SetStatusCode(e.GetChainStatusCode())

	writeErrorAsJSON(e, ctx.Response.BodyWriter())
}

func unwrapError(e *Error) *Error {
	if e != nil && e.Err != nil {
		if ee, ok := e.Err.(*Error); ok { // nolint: errorlint
			return ee
		}
	}

	return nil
}
