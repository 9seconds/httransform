package http

import (
	"context"
	"io"
	"net/http/httputil"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/valyala/fasthttp"
)

// Execute sends an http request and assign a streaming body to the
// given response. conn is a closable connection to the netloc.
func Execute(ctx context.Context, // nolint: funlen, cyclop
	conn io.ReadWriteCloser,
	request *fasthttp.Request,
	response *fasthttp.Response) error {
	subCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-subCtx.Done():
		case <-ctx.Done():
			select {
			case <-subCtx.Done():
			default:
				conn.Close()
			}
		}
	}()

	if _, err := request.WriteTo(conn); err != nil {
		return &errors.Error{
			Message: "cannot send a request",
			Code:    "http",
			Err:     err,
		}
	}

	response.Reset()
	response.Header.DisableNormalizing()

	bufReader := acquireBufioReader(conn)

	for code := fasthttp.StatusContinue; code == fasthttp.StatusContinue; code = response.Header.StatusCode() {
		if err := response.Header.Read(bufReader); err != nil {
			releaseBufioReader(bufReader)

			return &errors.Error{
				Message: "cannot read response headers",
				Code:    "http",
				Err:     err,
			}
		}
	}

	contentLength := response.Header.ContentLength()

	switch {
	case contentLength == 0 || request.Header.IsHead():
		response.SkipBody = true
	case contentLength > 0:
		reader := &closingReader{
			bufReader: bufReader,
			reader:    io.LimitReader(bufReader, int64(contentLength)),
		}
		response.SetBodyStream(reader, contentLength)
	default:
		reader := &closingReader{
			bufReader: bufReader,
			reader:    httputil.NewChunkedReader(bufReader),
		}
		response.SetBodyStream(reader, -1)
	}

	return nil
}
