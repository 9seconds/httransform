package client

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

const (
	DefaultHTTPTImeout = 3 * time.Minute
	DefaultHTTPPort    = "80"
	DefaultHTTPSPort   = "443"
)

type Client struct {
	dialer    Dialer
	tlsConfig *tls.Config
}

func (c *Client) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return c.do(req, resp, timeout, timeout)
}

func (c *Client) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.do(req, resp, DefaultHTTPTImeout, DefaultHTTPTImeout)
}

func (c *Client) do(req *fasthttp.Request, resp *fasthttp.Response, readTimeout, writeTimeout time.Duration) error {
	uri := req.URI()
	addr := string(uri.Host())
	isHTTP := bytes.Equal(bytes.ToLower(uri.Scheme()), []byte("http"))

	if _, _, err := net.SplitHostPort(addr); err != nil {
		if isHTTP {
			addr = net.JoinHostPort(addr, DefaultHTTPPort)
		} else {
			addr = net.JoinHostPort(addr, DefaultHTTPSPort)
		}
	}

	conn, err := c.dialer.Dial(addr)
	if err != nil {
		return errors.Annotate(err, "Cannot dial to addr")
	}
	if !isHTTP {
		conn = tls.Client(conn, c.tlsConfig)
	}

	if err = conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		conn.Close()
		c.dialer.NotifyClosed(addr)
		return errors.Annotate(err, "Cannot set write deadline")
	}
	if _, err = req.WriteTo(conn); err != nil {
		conn.Close()
		c.dialer.NotifyClosed(addr)
		return errors.Annotate(err, "Cannot send request")
	}

	resp.Reset()
	resp.Header.DisableNormalizing()

	if err = conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		conn.Close()
		c.dialer.NotifyClosed(addr)
		return errors.Annotate(err, "Cannot set read deadline")
	}

	connReader := poolBufferedReader.Get().(*bufio.Reader)
	connReader.Reset(conn)

	if err = resp.Header.Read(connReader); err != nil {
		return errors.Annotate(err, "Cannot read response header")
	}
	if resp.SkipBody {
		c.dialer.Release(conn, addr)
		return nil
	}

	var reader io.ReadCloser
	if resp.Header.ContentLength() >= 0 {
		reader = newSimpleReader(addr,
			conn,
			connReader,
			c.dialer,
			resp.Header.ConnectionClose(),
			int64(resp.Header.ContentLength()))
	} else {
		reader = newChunkedReader(addr,
			conn,
			connReader,
			c.dialer,
			resp.Header.ConnectionClose())
	}
	resp.SetBodyStream(reader, -1)

	return nil
}

func NewClient(dialer Dialer, tlsConfig *tls.Config) *Client {
	return &Client{
		dialer:    dialer,
		tlsConfig: tlsConfig,
	}
}
