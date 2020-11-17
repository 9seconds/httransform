package client

import (
	"bufio"
	"context"
	"io"
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
