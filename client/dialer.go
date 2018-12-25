package client

import (
	"net"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	// DefaultDialTimeout is a  timeout for dialing to the host
	DefaultDialTimeout = 20 * time.Second

	// DefaultLimit constraints an amount of connections to the hostname
	// for PooledDialer.
	DefaultLimit = 1024
)

// BaseDialer connects to the hostname.
type BaseDialer func(string, time.Duration) (net.Conn, error)

// Dialer is a common interface for data structure which dials to
// hostnames. It's main purpose is to release connections when necessary
// or notify that some net.Conn is broken.
type Dialer interface {
	// Dial establishes connection to the given host.
	Dial(string) (net.Conn, error)

	// Release returns net.Conn back to the client. Second parameter is
	// hostname which was used for Dial.
	Release(net.Conn, string)

	// NotifyClosed tells Dialer that net.Conn for address is broken.
	NotifyClosed(string)
}

// FastHTTPBaseDialer is a BaseDialer which uses fasthttp.
func FastHTTPBaseDialer(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return fasthttp.DialDualStackTimeout(addr, dialTimeout)
}

// StdBaseDialer is a BaseDialer which uses net.Dial.
func StdBaseDialer(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return (&net.Dialer{
		Timeout:   dialTimeout,
		DualStack: true,
	}).Dial("tcp", addr)
}
