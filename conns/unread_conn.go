package conns

import (
	"bytes"
	"io"
	"net"
	"sync"
)

// UnreadConn can 'unread' bytes which were already read from the
// connection. It has 3 modes: exploring, unreading and sealed.
//
// By default UnreadConn starts in exploring mode. Each byte which is
// read from underlying conn is stored in a buffer.
//
// When you tranfer UnreadConn into unreading mode, it flushes this
// internal buffer first and continue to read from the connection.
// Effectively it means that you re-read same bytes.
//
// Sealed means that internal buffer is flushed and we read from the
// connection directly.
type UnreadConn struct {
	net.Conn

	mutex  sync.RWMutex
	buf    bytes.Buffer
	reader io.Reader
}

// Read to conform io.Reader interface.
func (u *UnreadConn) Read(p []byte) (int, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	return u.reader.Read(p) // nolint: wrapcheck
}

// Unread transitions UnreadConn into 'unreading' state.
func (u *UnreadConn) Unread() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.reader = io.MultiReader(&u.buf, u.Conn)
}

// Seal transitions UnreadConn into 'sealed' state.
func (u *UnreadConn) Seal() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.buf.Reset()
	u.reader = u.Conn
}

// NewUnreadConn creates a new UnreadConn based on given net.Conn
// instance.
func NewUnreadConn(conn net.Conn) *UnreadConn {
	rv := &UnreadConn{
		Conn: conn,
	}
	rv.reader = io.TeeReader(conn, &rv.buf)

	return rv
}
