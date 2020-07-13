package client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/PumpkinSeed/errors"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/client/connectors"
	"github.com/9seconds/httransform/v2/client/readers"
)

type Client struct {
	connector       connectors.Connector
	tlsConfigs      map[string]*tls.Config
	tlsConfigsMutex sync.Mutex
}

func (c *Client) DoTimeout(ctx context.Context, request *fasthttp.Request, response *fasthttp.Response, timeout time.Duration) error {
	if timeout == 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Do(ctx, request, response)
}

func (c *Client) Do(ctx context.Context, request *fasthttp.Request, response *fasthttp.Response) error {
	// Saving and restoring of RequestURI is yet another way how to bypass
	// fasthttp which SUDDENLY thinks that it is quite smart to parse host
	// and mangle requestURI content here.
	//
	// If you set full URI there and execute URI() (READ-ONLY!!!) then
	// fasthttp will parse it in background.
	//
	// Godlike design.
	originalURI := request.Header.RequestURI()
	uri := request.URI()
	addr := string(uri.Host())
	isHTTP := bytes.EqualFold(uri.Scheme(), []byte("http"))

	if _, _, err := net.SplitHostPort(addr); err != nil {
		if isHTTP {
			addr = net.JoinHostPort(addr, DefaultHTTPPort)
		} else {
			addr = net.JoinHostPort(addr, DefaultHTTPSPort)
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	request.SetRequestURIBytes(originalURI)

	conn, err := c.connector.Connect(ctx, addr)
	if err != nil {
		return errors.Wrap(err, ErrClient)
	}

	if isHTTP {
		conn = connectors.NewTLSConn(conn, c.getTLSConfig(addr))
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	if _, err := request.WriteTo(conn); err != nil {
		return errors.Wrap(err, ErrClient)
	}

	response.Reset()
	response.Header.DisableNormalizing()

	connReader := bufio.NewReader(conn)
	if err := response.Header.Read(connReader); err != nil {
		return errors.Wrap(err, ErrClient)
	}

	if request.Header.IsHead() {
		conn.Release()
		return nil
	}

	contentLength := response.Header.ContentLength()
	if contentLength >= 0 {
		reader := readers.NewSimpleReader(conn,
			connReader,
			response.Header.ConnectionClose(),
			int64(contentLength))
		response.SetBodyStream(reader, contentLength)
		return nil
	}

	reader := readers.NewChunkedReader(conn,
		connReader,
		response.Header.ConnectionClose())
	response.SetBodyStream(reader, -1)

	return nil
}

func (c *Client) getTLSConfig(addr string) *tls.Config {
	if conf, ok := c.tlsConfigs[addr]; ok {
		return conf
	}

	c.tlsConfigsMutex.Lock()
	defer c.tlsConfigsMutex.Unlock()

	if conf, ok := c.tlsConfigs[addr]; ok {
		return conf
	}

	serverName := addr

	if strings.Contains(addr, ":") {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			serverName = "*"
		} else {
			serverName = host
		}
	}

	conf := &tls.Config{
		ClientSessionCache: tls.NewLRUClientSessionCache(0),
		ServerName:         serverName,
	}
	c.tlsConfigs[addr] = conf

	return conf
}

func NewClient(connector connectors.Connector) *Client {
	return &Client{
		connector:  connector,
		tlsConfigs: map[string]*tls.Config{},
	}
}
