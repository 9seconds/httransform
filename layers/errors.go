package layers

import "github.com/valyala/fasthttp"

type Error struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *Error) Error() string {
	return e.Err.Error() + ": " + e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) GetStatusCode() int {
	if e.StatusCode == 0 {
		return fasthttp.StatusInternalServerError
	}

	return e.StatusCode
}
