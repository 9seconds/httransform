package client

import (
	"bufio"
	"context"
	"io"
	"sync"
)

const (
	BufferedConnSize = 32 * 1024
)

type bufferedConn struct {
	rd     *bufio.Reader
	cancel context.CancelFunc
}

func (b *bufferedConn) Read(p []byte) (int, error) {
	return b.rd.Read(p)
}

func (b *bufferedConn) Cancel() {
	b.cancel()
}

func (b *bufferedConn) Reset(rd io.Reader, cancel context.CancelFunc) {
	b.rd.Reset(rd)
	b.cancel = cancel
}

var poolBufferedConn = sync.Pool{
	New: func() interface{} {
		return &bufferedConn{
			rd: bufio.NewReaderSize(nil, BufferedConnSize),
		}
	},
}

func acquireBufferedConn(rd io.Reader, cancel context.CancelFunc) *bufferedConn {
	reader := poolBufferedConn.Get().(*bufferedConn)

	reader.Reset(rd, cancel)

	return reader
}

func releaseBufferedConn(conn *bufferedConn) {
	conn.Reset(nil, nil)
	poolBufferedConn.Put(conn)
}
