package dialer

import (
	"net"
	"time"
)

type SimpleDialer struct {
	base    BaseDialer
	timeout time.Duration
}

func (s *SimpleDialer) Dial(addr string) (net.Conn, error) {
	return s.base(addr, s.timeout)
}

func (s *SimpleDialer) Release(conn net.Conn, _ string) {
	conn.Close()
}

func (s *SimpleDialer) NotifyClosed(_ string) {
}

func NewSimpleDialer(base BaseDialer, timeout time.Duration) (*SimpleDialer, error) {
	return &SimpleDialer{
		base:    base,
		timeout: timeout,
	}, nil
}
