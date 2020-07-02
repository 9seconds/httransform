package client2

import (
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/client2/dialers"
)

const (
	DefaultHTTPTimeout = 3 * time.Minute

	DefaultHTTPPort  = "80"
	DefaultHTTPSPort = "443"
)

type Client struct {
	dialer     dialers.Dialer
	tlsConfigs map[string]*tls.Config
	mutex      sync.RWMutex
}

func (c *Client) DoTimeout(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	if timeout == 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Do(ctx, req, resp)
}

func (c *Client) Do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error {
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

	conn, err := c.dialer.Dial(ctx, addr)
	if err != nil {
		return ErrClientCannotDial.WrapErr(err)
	}
}
