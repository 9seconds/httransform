package readers

import (
	"bufio"
	"io"
	"sync"

	"github.com/9seconds/httransform/v2/client2/dialers"
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
	b.closeOnce.Do(func() {
		b.closed = true
		b.conn.Close()
	})

	return nil
}

func (b *baseReader) Release() {
	if b.forceClose {
		b.Close()
	} else {
		b.closeOnce.Do(func() {
			b.closed = true
			b.conn.Release()
		})
	}
}
