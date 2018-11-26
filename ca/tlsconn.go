package ca

import (
	"net"
	"time"
)

type Release func()

type tlsConn struct {
	conn     net.Conn
	released bool
	release  Release
}

func (t *tlsConn) Read(b []byte) (int, error) {
	return t.conn.Read(b)
}

func (t *tlsConn) Write(b []byte) (int, error) {
	return t.conn.Write(b)
}

func (t *tlsConn) Close() error {
	if !t.released {
		t.release()
	}
	return t.conn.Close()
}

func (t *tlsConn) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

func (t *tlsConn) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

func (t *tlsConn) SetDeadline(tm time.Time) error {
	return t.conn.SetDeadline(tm)
}

func (t *tlsConn) SetReadDeadline(tm time.Time) error {
	return t.conn.SetReadDeadline(tm)
}

func (t *tlsConn) SetWriteDeadline(tm time.Time) error {
	return t.conn.SetWriteDeadline(tm)
}

func NewTLSConn(conn net.Conn, release Release) net.Conn {
	return &tlsConn{
		conn:    conn,
		release: release,
	}
}
