package client

import (
	"net"
	"time"
)

// SimpleDialer is a dialer which always closes connections when they
// are not needed. It does not do any pooling.
type SimpleDialer struct {
	base    BaseDialer
	timeout time.Duration
}

// Dial establishes connection to the given host.
func (s *SimpleDialer) Dial(addr string) (net.Conn, error) {
	return s.base(addr, s.timeout)
}

// Release returns net.Conn back to the client. Second parameter is
// hostname which was used for Dial.
func (s *SimpleDialer) Release(conn net.Conn, _ string) {
	conn.Close()
}

// NotifyClosed tells Dialer that net.Conn for address is broken.
func (s *SimpleDialer) NotifyClosed(_ string) {
}

// NewSimpleDialer returns new instance of SimpleDialer.
func NewSimpleDialer(base BaseDialer, timeout time.Duration) (*SimpleDialer, error) {
	return &SimpleDialer{
		base:    base,
		timeout: timeout,
	}, nil
}
