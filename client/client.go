package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/libp2p/go-reuseport"
	"github.com/valyala/fasthttp"
)

type Client struct {
	dialer Dialer
}

func (c *Client) Do(ctx context.Context, request *fasthttp.Request, response *fasthttp.Response) error {
	uri := request.URI()
	host := string(uri.Host())
	isPlain := bytes.EqualFold(uri.Scheme(), []byte("http"))

	var addrError *net.AddrError

	hostname, _, err := net.SplitHostPort(host)
	switch {
	case errors.As(err, &addrError) && strings.Contains(addrError.Err, "missing port"):
		hostname = host
		if isPlain {
			uri.SetHost(net.JoinHostPort(host, "80"))
		} else {
			uri.SetHost(net.JoinHostPort(host, "443"))
		}
	case err != nil:
		return fmt.Errorf("incorrect host %s: %w", host, err)
	}

	fmt.Println(hostname, request.URI().String())

	return nil
}

func NewClient(ctx context.Context) *Client {
	rv := &Client{
		dialer: Dialer{
			Dialer: net.Dialer{
				Control: reuseport.Control,
			},
		},
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rv.dialer.dnsCache.Refresh(true)
			}
		}
	}()

	return rv
}
