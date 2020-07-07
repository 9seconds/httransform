package readers

import (
	"bufio"
	"io"
	"sync"

	"github.com/9seconds/httransform/v2/client3/dialers"
)

type baseReader struct {
	reader     *bufio.Reader
	conn       dialers.Conn
	closeOnce  sync.Once
	forceClose bool
	closed     bool
	bytesLeft  int64
}

func (b *baseReader) Read(p []byte) (int, error) {
	if b.bytesLeft <= 0 {
		return 0, io.EOF
	}

	if int64(len(p)) > b.bytesLeft {
		p = p[:b.bytesLeft]
	}

	n, err := b.reader.Read(p)
	b.bytesLeft -= int64(n)

	return n, err
}

func (b *baseReader) Close() error {
	var err error

	b.closeOnce.Do(func() {
		b.closed = true
		err = b.conn.Close()
	})

	return err
}

func (b *baseReader) Release() {
	b.closeOnce.Do(func() {
		b.closed = true
		if b.forceClose {
			b.conn.Close()
		} else {
			b.conn.Release()
		}
	})
}
