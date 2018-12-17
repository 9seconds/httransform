package client

import (
	"net"
	"time"
)

type releasableConnection struct {
	conn net.Conn
	addr string
}

func (r releasableConnection) Read(p []byte) (int, error) {
	return r.conn.Read(p)
}

func (r releasableConnection) Write(p []byte) (int, error) {
	return r.conn.Write(p)
}

func (r releasableConnection) Close() error {
	return r.conn.Close()
}

func (r releasableConnection) LocalAddr() net.Addr {
	return r.conn.LocalAddr()
}

func (r releasableConnection) RemoteAddr() net.Addr {
	return r.conn.RemoteAddr()
}

func (r releasableConnection) SetDeadline(t time.Time) error {
	return r.conn.SetDeadline(t)
}

func (r releasableConnection) SetReadDeadline(t time.Time) error {
	return r.conn.SetReadDeadline(t)
}

func (r releasableConnection) SetWriteDeadline(t time.Time) error {
	return r.conn.SetWriteDeadline(t)
}
