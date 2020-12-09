package layers

import "github.com/valyala/fasthttp"

// Error defines a custom error which can be returned from a layer. If
// layer returns this error, the rest of machinery can use it to specify
// some status codes or content types.
type Error struct {
	// StatusCode defines a status code which should be returned to a
	// client.
	StatusCode int

	// ContentType defines a content type which should be used in a
	// response to a customer.
	ContentType string

    // Details defines some additional details for the error
	Details     string

    // Err keeps an original error.
	Err         error
}

// Error to conform error interface.
func (e *Error) Error() string {
	return e.Err.Error() + ": " + e.Details
}

// Unwrap to conform go1.13 error interface.
func (e *Error) Unwrap() error {
	return e.Err
}

// GetStatusCode returns a status code for the error.
func (e *Error) GetStatusCode() int {
	if e.StatusCode == 0 {
		return fasthttp.StatusInternalServerError
	}

	return e.StatusCode
}

// GetContentType returns a content type for the error.
func (e *Error) GetContentType() string {
	if e.ContentType == "" {
		return "text/plain"
	}

	return e.ContentType
}
