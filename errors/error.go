package errors

import (
	"github.com/valyala/fasthttp"
)

// Error defines a custom error which can be returned from a layer or
// executor. This error has a stack of attached errors and can render
// JSON. Also, it keeps a status code so if you are searching for a
// correct way of setting your own response code, this is a best one.
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

func (e *Error) Unwrap() error {
	if e != nil {
		return e.Err
	}

	return nil
}

func (e *Error) GetStatusCode() int {
	if e != nil && e.StatusCode != 0 {
		return e.StatusCode
	}

	return 0
}

func (e *Error) GetChainStatusCode() int {
	for current := e; current != nil; current = unwrapError(current) {
		if current.StatusCode != 0 {
			return current.StatusCode
		}
	}

	return fasthttp.StatusInternalServerError
}

func (e *Error) GetMessage() string {
	if e != nil {
		return e.Message
	}

	return ""
}

func (e *Error) GetCode() string {
	if e != nil {
		return e.Code
	}

	return ""
}

func (e *Error) GetChainCode() string {
	for current := e; current != nil; current = unwrapError(current) {
		if e.Code != "" {
			return e.Code
		}
	}

	return "internal_error"
}

func (e *Error) ErrorJSON() string {
	encoder := acquireErrorEncoder()
	defer releaseErrorEncoder(encoder)

	encoder.obj.Error.Code = e.GetChainCode()
	encoder.obj.Error.Message = e.GetMessage()
	lastMessage := ""

	for current := e; current != nil; current = unwrapError(current) {
		encoder.obj.Error.Stack = append(encoder.obj.Error.Stack, errorJSONStackEntry{
			Message:    current.Message,
			Code:       current.Code,
			StatusCode: current.StatusCode,
		})

		if current.Err != nil {
			if _, ok := current.Err.(*Error); !ok { // nolint: errorlint
				lastMessage = current.Err.Error()
			}
		}
	}

	if lastMessage != "" {
		encoder.obj.Error.Stack = append(encoder.obj.Error.Stack, errorJSONStackEntry{
			Message: lastMessage,
		})
	}

	return encoder.Encode()
}

func (e *Error) WriteTo(ctx *fasthttp.RequestCtx) {
	ctx.Response.Reset()
	ctx.Response.Header.DisableNormalizing()
	ctx.Response.SetConnectionClose()
	ctx.Response.Header.SetContentType("application/json")
	ctx.Response.Header.SetStatusCode(e.GetChainStatusCode())
	ctx.Response.SetBodyString(e.ErrorJSON())
}

func unwrapError(e *Error) *Error {
	if e != nil && e.Err != nil {
		if ee, ok := e.Err.(*Error); ok { // nolint: errorlint
			return ee
		}
	}

	return nil
}
