package client

import (
	"net"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	DefaultDialTimeout = 20 * time.Second
	DefaultLimit       = 1024
)

type BaseDialer func(string, time.Duration) (net.Conn, error)

type Dialer interface {
	Dial(string) (net.Conn, error)
	Release(net.Conn, string)
	NotifyClosed(string)
}

func FastHTTPBaseDialer(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return fasthttp.DialDualStackTimeout(addr, dialTimeout)
}

func StdBaseDialer(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return (&net.Dialer{
		Timeout:   dialTimeout,
		DualStack: true,
	}).Dial("tcp", addr)
}
