package dialers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"github.com/PumpkinSeed/errors"
	"github.com/valyala/fasthttp"
)

type connectDialer struct {
	dialer     Dialer
	address    string
	authHeader string
}

func (c *connectDialer) Dial(ctx context.Context, addr string) (net.Conn, error) {
	conn, err := c.dialer.Dial(ctx, c.address)
	if err != nil {
		return nil, errors.Wrap(err, ErrDialer)
	}

	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "443")
	}

	buf := acquireConnectRequestByteBuffer()
	defer releaseConnectRequestByteBuffer(buf)

	fmt.Fprintf(buf, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)

	if c.authHeader != "" {
		fmt.Fprintf(buf, "Proxy-Authorization: %s\r\n", c.authHeader)
	}

	buf.WriteString("\r\n")

	if _, err := conn.Write(buf.Bytes()); err != nil {
		conn.Close()
		return nil, errors.Wrap(err, ErrCannotConnectToProxy)
	}

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	response.SkipBody = true

	reader := acquireConnectResponseBufioReader(conn)
	defer releaseConnectResponseBufioReader(reader)

	if err := response.Read(reader); err != nil {
		conn.Close()
		return nil, errors.Wrap(err, ErrCannotReadResponse)
	}

	if response.StatusCode() != fasthttp.StatusOK {
		conn.Close()
		return nil, errors.Wrap(fmt.Errorf("response status code is %d", response.StatusCode()), ErrProxyRejected)
	}

	return conn, nil
}

func (c *connectDialer) GetTimeout() time.Duration {
	return c.dialer.GetTimeout()
}

func NewConnectDialer(address string, timeout time.Duration, user, password string) Dialer {
	proxyAuthHeader := ""
	if user != "" || password != "" {
		proxyAuthHeader = base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
	}

	return &connectDialer{
		dialer:     NewSimpleDialer(timeout),
		address:    address,
		authHeader: proxyAuthHeader,
	}
}
