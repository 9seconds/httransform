package client

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

const (
	// DefaultHTTPTimeout is a default timeout for processing of a single
	// HTTP request.
	DefaultHTTPTimeout = 3 * time.Minute

	// DefaultHTTPPort is a default port for HTTP.
	DefaultHTTPPort = "80"

	// DefaultHTTPSPort is a default port for HTTPS.
	DefaultHTTPSPort = "443"
)

// Client is the implementation of HTTP1 client which sets body as a
// stream.
type Client struct {
	dialer           Dialer
	tlsConfigMap     map[string]*tls.Config
	tlsConfigMapLock sync.Mutex
}

// DoTimeout does HTTP request with the given timeout.
func (c *Client) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return c.do(req, resp, timeout, timeout)
}

// Do does HTTP request with the default timeout.
func (c *Client) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.do(req, resp, DefaultHTTPTimeout, DefaultHTTPTimeout)
}

func (c *Client) do(req *fasthttp.Request, resp *fasthttp.Response, readTimeout, writeTimeout time.Duration) error { // nolint: funlen
	// Saving and restoring of RequestURI is yet another way how to bypass
	// fasthttp which SUDDENLY thinks that it is quite smart to parse host
	// and mangle requestURI content here.
	//
	// If you set full URI there and execute URI() (READ-ONLY!!!) then
	// fasthttp will parse it in background.
	//
	// Godlike design.
	originalURI := req.Header.RequestURI()
	uri := req.URI()
	addr := string(uri.Host())
	isHTTP := bytes.EqualFold(uri.Scheme(), []byte("http"))

	if _, _, err := net.SplitHostPort(addr); err != nil {
		if isHTTP {
			addr = net.JoinHostPort(addr, DefaultHTTPPort)
		} else {
			addr = net.JoinHostPort(addr, DefaultHTTPSPort)
		}
	}

	req.SetRequestURIBytes(originalURI)

	conn, err := c.dialer.Dial(addr)
	if err != nil {
		return xerrors.Errorf("cannot dial to addr: %w", err)
	}

	if !isHTTP {
		conn = tls.Client(conn, c.cachedTLSConfig(addr))
	}

	if err = conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		conn.Close()
		c.dialer.NotifyClosed(addr)

		return xerrors.Errorf("cannot set write deadline: %w", err)
	}

	if _, err = req.WriteTo(conn); err != nil {
		conn.Close()
		c.dialer.NotifyClosed(addr)

		return xerrors.Errorf("cannot send a request: %w", err)
	}

	resp.Reset()
	resp.Header.DisableNormalizing()

	if err = conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		conn.Close()
		c.dialer.NotifyClosed(addr)

		return xerrors.Errorf("cannot set read deadline: %w", err)
	}

	connReader := poolBufferedReader.Get().(*bufio.Reader)
	connReader.Reset(conn)

	if err = resp.Header.Read(connReader); err != nil {
		return xerrors.Errorf("cannot read response header: %w", err)
	}

	if req.Header.IsHead() {
		c.dialer.Release(conn, addr)
		return nil
	}

	var reader io.ReadCloser

	contentLength := resp.Header.ContentLength()

	if contentLength >= 0 {
		reader = newSimpleReader(addr,
			conn,
			connReader,
			c.dialer,
			resp.Header.ConnectionClose(),
			int64(contentLength))
		resp.SetBodyStream(reader, contentLength)
	} else {
		reader = newChunkedReader(addr,
			conn,
			connReader,
			c.dialer,
			resp.Header.ConnectionClose())
		resp.SetBodyStream(reader, -1)
	}

	return nil
}

func (c *Client) cachedTLSConfig(addr string) *tls.Config {
	c.tlsConfigMapLock.Lock()

	cfg := c.tlsConfigMap[addr]
	if cfg == nil {
		cfg = newClientTLSConfig(addr)
		c.tlsConfigMap[addr] = cfg
	}
	c.tlsConfigMapLock.Unlock()

	return cfg
}

func newClientTLSConfig(addr string) *tls.Config {
	c := &tls.Config{}
	if c.ClientSessionCache == nil {
		c.ClientSessionCache = tls.NewLRUClientSessionCache(0)
	}

	if len(c.ServerName) == 0 {
		serverName := tlsServerName(addr)
		if serverName == "*" {
			c.InsecureSkipVerify = true
		} else {
			c.ServerName = serverName
		}
	}

	return c
}

func tlsServerName(addr string) string {
	if !strings.Contains(addr, ":") {
		return addr
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "*"
	}

	return host
}

// NewClient creates a new instance of HTTP1 client.
func NewClient(dialer Dialer) *Client {
	return &Client{
		dialer:       dialer,
		tlsConfigMap: make(map[string]*tls.Config),
	}
}
