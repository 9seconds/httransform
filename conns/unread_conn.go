package conns

import (
	"bytes"
	"io"
	"net"
	"sync"
)

type UnreadConn struct {
	net.Conn

	mutex           sync.RWMutex
	buf             bytes.Buffer
	exploringReader io.Reader
	sealedReader    io.Reader
	exploring       bool
}

func (u *UnreadConn) Read(p []byte) (int, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	if u.exploring {
		return u.exploringReader.Read(p)
	}

	return u.sealedReader.Read(p)
}

func (u *UnreadConn) Unread() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.sealedReader = u.Conn
	u.exploring = false
}

func (u *UnreadConn) Seal() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.exploring = false
}

func NewUnreadConn(conn net.Conn) *UnreadConn {
	rv := &UnreadConn{
		Conn:      conn,
		exploring: true,
	}
	rv.sealedReader = io.TeeReader(conn, &rv.buf)
	rv.exploringReader = io.MultiReader(&rv.buf, conn)

	return rv
}
