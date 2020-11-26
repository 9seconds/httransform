package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http/httputil"

	"github.com/valyala/fasthttp"
)

func Execute(ctx context.Context,
	conn net.Conn,
	request *fasthttp.Request,
	response *fasthttp.Response) error {
	ctx, cancel := context.WithCancel(ctx)

	if _, err := request.WriteTo(conn); err != nil {
		cancel()

		return fmt.Errorf("cannot send a rquest: %w", err)
	}

	response.Reset()
	response.Header.DisableNormalizing()

	bufReader := acquireBufferedReader(conn, cancel)
	statusCode := fasthttp.StatusContinue

	for statusCode == fasthttp.StatusContinue {
		if err := response.Header.Read(bufReader.reader); err != nil {
			releaseBufferedReader(bufReader)
			cancel()

			return fmt.Errorf("cannot read response headers: %w", err)
		}

		statusCode = response.Header.StatusCode()
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
