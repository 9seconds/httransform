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
	baseDialer      *base
	connectRequest  []byte
	proxyHost       string
	proxyPort       string
	bufioReaderPool sync.Pool
}

func (h *httpProxy) Dial(ctx context.Context, host, port string) (net.Conn, error) {
	return h.baseDialer.Dial(ctx, h.proxyHost, h.proxyPort)
}

func (h *httpProxy) UpgradeToTLS(ctx context.Context, conn net.Conn, host string) (net.Conn, error) {
	ownCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-ownCtx.Done():
		case <-ctx.Done():
			conn.Close()
		}
	}()

	if _, err := conn.Write(h.connectRequest); err != nil {
		return nil, fmt.Errorf("cannot send a connect request: %w", err)
	}

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	bufReader := h.acquireBufioReader(conn)
	defer h.releaseBufioReader(bufReader)

	if err := response.Read(bufReader); err != nil {
		return nil, fmt.Errorf("cannot read http response: %w", err)
	}

	if response.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("proxy has responsed with %d status code", response.StatusCode())
	}

	return h.baseDialer.UpgradeToTLS(ctx, conn, host)
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

func NewHTTPProxy(opt Opts, proxyAuth ProxyAuth) (Dialer, error) {
	baseDialer, err := NewBase(opt)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize a base dialer: %w", err)
	}

	connectRequest := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", proxyAuth.Address)

	if proxyAuth.HasCredentials() {
		rawLine := proxyAuth.Username + ":" + proxyAuth.Password
		encodedLine := base64.StdEncoding.EncodeToString([]byte(rawLine))
		connectRequest += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encodedLine)
	}

	connectRequest += "\r\n"

	host, port, _ := net.SplitHostPort(proxyAuth.Address)

	return &httpProxy{
		baseDialer:     baseDialer.(*base),
		connectRequest: []byte(connectRequest),
		proxyHost:      host,
		proxyPort:      port,
		bufioReaderPool: sync.Pool{
			New: func() interface{} {
				return bufio.NewReaderSize(nil, 5*1024) // nolint: gomnd
			},
		},
	}, nil
}
