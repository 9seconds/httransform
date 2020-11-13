package client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/9seconds/httransform/v2/dialer"
	lru "github.com/hashicorp/golang-lru"
	"github.com/valyala/fasthttp"
)

const (
	TLSConfigsCacheMaxSize = 128
	BufioReaderSize        = 16 * 1024
)

type Client struct {
	dialer         dialer.Dialer
	tlsConfigs     *lru.Cache
	tlsConfigsLock sync.Mutex
	timeout        time.Duration
	tlsNoVerify    bool
}

func (c *Client) Do(ctx context.Context, request *fasthttp.Request, response *fasthttp.Response) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)

	uri := request.URI()
	host := string(uri.Host())
	isPlain := bytes.EqualFold(uri.Scheme(), []byte("http"))

	var addrError *net.AddrError

	hostname, _, err := net.SplitHostPort(host)
	switch {
	case errors.As(err, &addrError) && strings.Contains(addrError.Err, "missing port"):
		hostname = host
		if isPlain {
			host = net.JoinHostPort(host, "80")
		} else {
			host = net.JoinHostPort(host, "443")
		}
		uri.SetHost(host)
	case err != nil:
		cancel()
		return fmt.Errorf("incorrect host %s: %w", host, err)
	}

	conn, err := c.dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		cancel()
		return fmt.Errorf("cannot dial to host %s: %w", host, err)
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	if !isPlain {
		tlsConn := tls.Client(conn, c.getTLSConfig(hostname))
		if err := tlsConn.Handshake(); err != nil {
			cancel()
			return fmt.Errorf("cannot establish tls handshake: %w", err)
		}
		conn = tlsConn
	}

	if _, err := request.WriteTo(conn); err != nil {
		cancel()
		return fmt.Errorf("cannot send a request: %w", err)
	}

	response.Reset()
	response.Header.EnableNormalizing()

	bufConn := bufio.NewReaderSize(conn, BufioReaderSize)
	statusCode := fasthttp.StatusContinue

	for statusCode == fasthttp.StatusContinue {
		if err := response.Header.Read(bufConn); err != nil {
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
		cancel()
	case contentLength > 0:
		reader := &streamReader{
			data:   io.LimitReader(bufConn, int64(contentLength)),
			cancel: cancel,
		}
		response.SetBodyStream(reader, contentLength)
	default:
		reader := &streamReader{
			data:   httputil.NewChunkedReader(bufConn),
			cancel: cancel,
		}
		response.SetBodyStream(reader, -1)
	}

	return nil
}

func (c *Client) getTLSConfig(host string) *tls.Config {
	if conf, ok := c.tlsConfigs.Get(host); ok {
		return conf.(*tls.Config)
	}

	c.tlsConfigsLock.Lock()
	defer c.tlsConfigsLock.Unlock()

	if conf, ok := c.tlsConfigs.Get(host); ok {
		return conf.(*tls.Config)
	}

	conf := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: c.tlsNoVerify,
	}
	c.tlsConfigs.Add(host, conf)

	return conf
}

func NewClient(dial dialer.Dialer, httpTimeout time.Duration, tlsNoVerify bool) *Client {
	cache, err := lru.New(TLSConfigsCacheMaxSize)
	if err != nil {
		panic(err)
	}

	rv := &Client{
		dialer:      dial,
		tlsConfigs:  cache,
		timeout:     httpTimeout,
		tlsNoVerify: tlsNoVerify,
	}

	return rv
}
