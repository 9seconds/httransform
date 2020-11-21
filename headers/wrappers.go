package headers

import (
	"bufio"
	"fmt"
	"io"

	"github.com/valyala/fasthttp"
)

type RequestHeaderWrapper struct {
	ref *fasthttp.RequestHeader
}

func (r RequestHeaderWrapper) Read(rd io.Reader) error {
	method := append([]byte(nil), r.ref.Method()...)
	requestURI := append([]byte(nil), r.ref.RequestURI()...)
	host := append([]byte(nil), r.ref.Host()...)

	r.ref.Reset()
	r.ref.DisableNormalizing()

	if err := r.ref.Read(bufio.NewReader(rd)); err != nil {
		return fmt.Errorf("cannot read request headers: %w", err)
	}

	r.ref.SetHostBytes(host)
	r.ref.SetMethodBytes(method)
	r.ref.SetRequestURIBytes(requestURI)

	return nil
}

func (r RequestHeaderWrapper) DisableNormalizing() {
	r.ref.DisableNormalizing()
}

func (r RequestHeaderWrapper) Headers() []byte {
	return r.ref.RawHeaders()
}

type ResponseHeaderWrapper struct {
	ref *fasthttp.ResponseHeader
}

func (r ResponseHeaderWrapper) Read(rd io.Reader) error {
	statusCode := r.ref.StatusCode()

	r.ref.Reset()
	r.ref.DisableNormalizing()

	if err := r.ref.Read(bufio.NewReader(rd)); err != nil {
		return fmt.Errorf("cannot read response headers: %w", err)
	}

	r.ref.SetStatusCode(statusCode)

	return nil
}

func (r ResponseHeaderWrapper) DisableNormalizing() {
	r.ref.DisableNormalizing()
}

func (r ResponseHeaderWrapper) Headers() []byte {
	return r.ref.Header()
}

func NewRequestHeaderWrapper(ref *fasthttp.RequestHeader) RequestHeaderWrapper {
	return RequestHeaderWrapper{ref}
}

func NewResponseHeaderWrapper(ref *fasthttp.ResponseHeader) ResponseHeaderWrapper {
	return ResponseHeaderWrapper{ref}
}
