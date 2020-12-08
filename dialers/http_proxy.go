package dialers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/valyala/fasthttp"
)

type httpProxy struct {
	baseDialer           *base
	connectRequestSuffix []byte
	proxyHost            string
	proxyPort            string
	bufioReaderPool      sync.Pool
	bytesBufferPool      sync.Pool
}

func (h *httpProxy) Dial(ctx context.Context, host, port string) (net.Conn, error) {
	return h.baseDialer.Dial(ctx, h.proxyHost, h.proxyPort)
}

func (h *httpProxy) UpgradeToTLS(ctx context.Context, conn net.Conn, host, port string) (net.Conn, error) {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		timer := fasthttp.AcquireTimer(h.baseDialer.netDialer.Timeout)
		defer fasthttp.ReleaseTimer(timer)

		select {
		case <-subCtx.Done():
		case <-ctx.Done():
			select {
			case <-subCtx.Done():
			default:
				conn.Close()
			}
		case <-timer.C:
			select {
			case <-subCtx.Done():
			default:
				conn.Close()
			}
		}
	}()

	buf := h.acquireBytesBuffer()
	defer h.releaseBytesBuffer(buf)

	buf.WriteString("CONNECT ")
	buf.WriteString(net.JoinHostPort(host, port))
	buf.WriteString(" HTTP/1.1\r\n")
	buf.Write(h.connectRequestSuffix)

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("cannot send a connect request: %w", err)
	}

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	response.SkipBody = true

	bufReader := h.acquireBufioReader(conn)
	defer h.releaseBufioReader(bufReader)

	if err := response.Read(bufReader); err != nil {
		return nil, fmt.Errorf("cannot read http response: %w", err)
	}

	if response.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("proxy has responsed with %d status code", response.StatusCode())
	}

	return h.baseDialer.UpgradeToTLS(subCtx, conn, host, port)
}

func (h *httpProxy) PatchHTTPRequest(req *fasthttp.Request) {
	if bytes.EqualFold(req.URI().Scheme(), []byte("http")) {
		req.SetRequestURIBytes(req.Header.RequestURI())
	}
}

func (h *httpProxy) acquireBufioReader(rd io.Reader) *bufio.Reader {
	rv := h.bufioReaderPool.Get().(*bufio.Reader)

	rv.Reset(rd)

	return rv
}

func (h *httpProxy) releaseBufioReader(reader *bufio.Reader) {
	reader.Reset(nil)
	h.bufioReaderPool.Put(reader)
}

func (h *httpProxy) acquireBytesBuffer() *bytes.Buffer {
	return h.bytesBufferPool.Get().(*bytes.Buffer)
}

func (h *httpProxy) releaseBytesBuffer(buf *bytes.Buffer) {
	buf.Reset()
	h.bytesBufferPool.Put(buf)
}

// NewHTTPProxy returns a dialer which dials using HTTP proxies It uses.
// a base dialer under the hood so you get all its niceties there      .
func NewHTTPProxy(opt Opts, proxyAuth ProxyAuth) Dialer {
	connectRequestSuffix := ""

	if proxyAuth.HasCredentials() {
		rawLine := proxyAuth.Username + ":" + proxyAuth.Password
		encodedLine := base64.StdEncoding.EncodeToString([]byte(rawLine))
		connectRequestSuffix += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encodedLine)
	}

	connectRequestSuffix += "\r\n"

	host, port, _ := net.SplitHostPort(proxyAuth.Address)

	return &httpProxy{
		baseDialer:           NewBase(opt).(*base),
		connectRequestSuffix: []byte(connectRequestSuffix),
		proxyHost:            host,
		proxyPort:            port,
		bufioReaderPool: sync.Pool{
			New: func() interface{} {
				return bufio.NewReaderSize(nil, 5*1024) // nolint: gomnd
			},
		},
		bytesBufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}
