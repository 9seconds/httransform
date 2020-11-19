package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/valyala/fasthttp"
)

const (
	BufferedConnSize = 32 * 1024
)

type Client struct {
	dialer      dialers.Dialer
	bufConnPool sync.Pool
}

func (c *Client) Do(ctx context.Context,
	connectAddress string,
	request *fasthttp.Request,
	response *fasthttp.Response) error {
	uri := request.URI()
	isSecure := bytes.EqualFold(uri.Scheme(), []byte("https"))

	conn, err := dialers.Dial(ctx, c.dialer, connectAddress, isSecure)
	if err != nil {
		return fmt.Errorf("cannot call to %s: %w", connectAddress, err)
	}

	if _, err := request.WriteTo(conn); err != nil {
		return fmt.Errorf("cannot send a request: %w", err)
	}

	response.Reset()
	response.Header.EnableNormalizing()

	bufConn := c.acquireBufferedConn(conn)
	statusCode := fasthttp.StatusContinue

	for statusCode == fasthttp.StatusContinue {
		if err := response.Header.Read(bufConn.rd); err != nil {
			c.releaseBufferedConn(bufConn)
			return fmt.Errorf("cannot read response headers: %w", err)
		}
	}

	response.SetConnectionClose()

	return nil
}

func (c *Client) acquireBufferedConn(rd io.Reader) *bufio.Reader {
	reader := c.bufConnPool.Get().(*bufio.Reader)

	reader.Reset(rd)

	return reader
}

func (c *Client) releaseBufferedConn(reader *bufio.Reader) {
	reader.Reset(nil)
	c.bufConnPool.Put(reader)
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

// func (c *Client) Do(ctx context.Context, request *fasthttp.Request, response *fasthttp.Response) error {
// 	uri := request.URI()
// 	isPlain := bytes.EqualFold(uri.Scheme(), []byte("http"))
// 	ctx, cancel := context.WithCancel(ctx)

// 	if err := c.ensureHostPort(request.URI(), isPlain); err != nil {
// 		return err
// 	}

// 	conn, err := c.dial(ctx, string(uri.Host()), isPlain)
// 	if err != nil {
// 		cancel()
// 		return err
// 	}

// 	go func() {
// 		<-ctx.Done()
// 		cancel()
// 		conn.Close()
// 	}()

// 	if _, err := request.WriteTo(conn); err != nil {
// 		cancel()
// 		return fmt.Errorf("cannot send a request: %w", err)
// 	}

// 	response.Reset()
// 	response.Header.EnableNormalizing()

// 	bufConn := acquireBufferedConn(conn, cancel)
// 	statusCode := fasthttp.StatusContinue

// 	for statusCode == fasthttp.StatusContinue {
// 		if err := response.Header.Read(bufConn.rd); err != nil {
// 			releaseBufferedConn(bufConn)
// 			cancel()
// 			return fmt.Errorf("cannot read response headers: %w", err)
// 		}
// 		statusCode = response.Header.StatusCode()
// 	}

// 	response.SetConnectionClose()

// 	contentLength := response.Header.ContentLength()

// 	switch {
// 	case contentLength == 0 || request.Header.IsHead():
// 		response.SkipBody = true
// 		cancel()
// 	case contentLength > 0:
// 		reader := &streamReader{
// 			bufferedConn: bufConn,
// 			toReadFrom:   io.LimitReader(bufConn, int64(contentLength)),
// 		}
// 		response.SetBodyStream(reader, contentLength)
// 	default:
// 		reader := &streamReader{
// 			bufferedConn: bufConn,
// 			toReadFrom:   httputil.NewChunkedReader(bufConn),
// 		}
// 		response.SetBodyStream(reader, -1)
// 	}

// 	return nil
// }

// func (c *Client) ensureHostPort(uri *fasthttp.URI, isPlain bool) error {
// 	host := string(uri.Host())

// 	var addrError *net.AddrError

// 	_, _, err := net.SplitHostPort(host)
// 	switch {
// 	case errors.As(err, &addrError) && strings.Contains(addrError.Err, "missing port"):
// 		if isPlain {
// 			host = net.JoinHostPort(host, "80")
// 		} else {
// 			host = net.JoinHostPort(host, "443")
// 		}
// 		uri.SetHost(host)
// 	case err != nil:
// 		return fmt.Errorf("incorrect host %s: %w", host, err)
// 	}

// 	return nil
// }

// func (c *Client) dial(ctx context.Context, hostport string, isPlain bool) (net.Conn, error) {
// 	conn, err := c.dialer.DialContext(ctx, "tcp", hostport)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot dial to host %s: %w", hostport, err)
// 	}

// 	host, _, _ := net.SplitHostPort(hostport)

// 	if !isPlain {
// 		tlsConn := tls.Client(conn, c.getTLSConfig(host))
// 		if err := tlsConn.Handshake(); err != nil {
// 			conn.Close()
// 			return nil, fmt.Errorf("cannot establish tls handshake: %w", err)
// 		}
// 		return tlsConn, nil
// 	}

// 	return conn, nil
// }

// func (c *Client) getTLSConfig(host string) *tls.Config {
// 	if conf, ok := c.tlsConfigs.Get(host); ok {
// 		return conf.(*tls.Config)
// 	}

// 	c.tlsConfigsLock.Lock()
// 	defer c.tlsConfigsLock.Unlock()

// 	if conf, ok := c.tlsConfigs.Get(host); ok {
// 		return conf.(*tls.Config)
// 	}

// 	conf := &tls.Config{
// 		ServerName:         host,
// 		InsecureSkipVerify: c.tlsNoVerify,
// 	}
// 	c.tlsConfigs.Add(host, conf)

// 	return conf
// }

// func NewClient(dial dialer.Dialer, tlsNoVerify bool) *Client {
// 	cache, err := lru.New(TLSConfigsCacheMaxSize)
// 	if err != nil {
// 		panic(err)
// 	}

// 	rv := &Client{
// 		dialer:      dial,
// 		tlsConfigs:  cache,
// 		tlsNoVerify: tlsNoVerify,
// 	}

// 	return rv
// }
