package dialers

import (
	"crypto/tls"
	"net"
	"time"
)

type tlsConn struct {
	parent Conn
	conn   tls.Conn
}

func (t *tlsConn) Read(p []byte) (int, error) {
	return t.conn.Read(p)
}

func (t *tlsConn) Write(p []byte) (int, error) {
	return t.conn.Write(p)
}

func (t *tlsConn) SetDeadline(tt time.Time) error {
	return t.conn.SetDeadline(tt)
}

func (t *tlsConn) SetReadDeadline(tt time.Time) error {
	return t.conn.SetReadDeadline(tt)
}

func (t *tlsConn) SetWriteDeadline(tt time.Time) error {
	return t.conn.SetWriteDeadline(tt)
}

func (t *tlsConn) Close() error {
	t.conn.Close()
	return t.parent.Close()
}

func (t *tlsConn) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

func (t *tlsConn) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

func (t *tlsConn) Release() {
	t.parent.Release()
}

func NewTLSConn(parent Conn, config *tls.Config) Conn {
	return &tlsConn{
		parent: parent,
		conn:   *tls.Client(parent, config),
	}
}
