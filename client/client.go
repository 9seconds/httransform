package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http/httputil"
	"sync"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/valyala/fasthttp"
)

type Client struct {
	dialer      dialers.Dialer
	bufConnPool sync.Pool
}

func (c *Client) Do(ctx context.Context, // nolint: funlen
	connectAddress string,
	request *fasthttp.Request,
	response *fasthttp.Response) error {
	uri := request.URI()
	isSecure := bytes.EqualFold(uri.Scheme(), []byte("https"))

	c.dialer.PatchHTTPRequest(request)

	conn, err := dialers.Dial(ctx, c.dialer, connectAddress, isSecure)
	if err != nil {
		return fmt.Errorf("cannot call to %s: %w", connectAddress, err)
	}

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	if _, err := request.WriteTo(conn); err != nil {
		cancel()

		return fmt.Errorf("cannot send a request: %w", err)
	}

	response.Reset()
	response.Header.DisableNormalizing()

	bufConn := acquireBufferedConn(conn, cancel)
	statusCode := fasthttp.StatusContinue

	for statusCode == fasthttp.StatusContinue {
		if err := response.Header.Read(bufConn.rd); err != nil {
			releaseBufferedConn(bufConn)
			cancel()

			return fmt.Errorf("cannot read response headers: %w", err)
		}

		statusCode = response.Header.StatusCode()
	}

	response.SetConnectionClose()

	contentLength := response.Header.ContentLength()

	switch {
	case contentLength == 0 || request.Header.IsHead():
		response.SkipBody = true
	case contentLength > 0:
		reader := &streamReader{
			bufferedConn: bufConn,
			toReadFrom:   io.LimitReader(bufConn, int64(contentLength)),
		}
		response.SetBodyStream(reader, contentLength)
	default:
		reader := &streamReader{
			bufferedConn: bufConn,
			toReadFrom:   httputil.NewChunkedReader(bufConn),
		}
		response.SetBodyStream(reader, -1)
	}

	return nil
}

func NewClient(dialer dialers.Dialer) *Client {
	return &Client{
		dialer: dialer,
		bufConnPool: sync.Pool{
			New: func() interface{} {
				return &bufferedConn{
					rd: bufio.NewReaderSize(nil, BufferedConnSize),
				}
			},
		},
	}
}
