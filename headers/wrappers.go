package headers

import (
	"io"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/valyala/fasthttp"
)

type requestHeaderWrapper struct {
	ref *fasthttp.RequestHeader
}

func (r requestHeaderWrapper) Read(rd io.Reader) error {
	method := append([]byte(nil), r.ref.Method()...)
	requestURI := append([]byte(nil), r.ref.RequestURI()...)
	host := append([]byte(nil), r.ref.Host()...)

	r.ref.Reset()
	r.ref.DisableNormalizing()

	bufReader := acquireBufioReader(rd)
	defer releaseBufioReader(bufReader)

	if err := r.ref.Read(bufReader); err != nil {
		return errors.Annotate(err, "cannot read request headers", "headers_sync", 0)
	}

	r.ref.SetHostBytes(host)
	r.ref.SetMethodBytes(method)
	r.ref.SetRequestURIBytes(requestURI)

	return nil
}

func (r requestHeaderWrapper) DisableNormalizing() {
	r.ref.DisableNormalizing()
}

func (r requestHeaderWrapper) ResetConnectionClose() {
	r.ref.ResetConnectionClose()
}

func (r requestHeaderWrapper) Headers() []byte {
	return r.ref.RawHeaders()
}

type responseHeaderWrapper struct {
	ref *fasthttp.ResponseHeader
}

func (r responseHeaderWrapper) Read(rd io.Reader) error {
	statusCode := r.ref.StatusCode()

	r.ref.Reset()
	r.ref.DisableNormalizing()

	bufReader := acquireBufioReader(rd)
	defer releaseBufioReader(bufReader)

	if err := r.ref.Read(bufReader); err != nil {
		return errors.Annotate(err, "cannot read response headers", "headers_sync", 0)
	}

	r.ref.SetStatusCode(statusCode)

	return nil
}

func (r responseHeaderWrapper) DisableNormalizing() {
	r.ref.DisableNormalizing()
}

func (r responseHeaderWrapper) ResetConnectionClose() {
	r.ref.ResetConnectionClose()
}

func (r responseHeaderWrapper) Headers() []byte {
	return r.ref.Header()
}

// NewRequestHeaderWrapper returns a new instance of
// requestHeaderWrapper for a given fasthttp.RequestHeader.
func NewRequestHeaderWrapper(ref *fasthttp.RequestHeader) FastHTTPHeaderWrapper {
	return requestHeaderWrapper{ref}
}

// NewResponseHeaderWrapper returns a new instance of
// responseHeaderWrapper for a given fasthttp.ResponseHeader.
func NewResponseHeaderWrapper(ref *fasthttp.ResponseHeader) FastHTTPHeaderWrapper {
	return responseHeaderWrapper{ref}
}
