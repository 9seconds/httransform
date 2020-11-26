package http

import (
	"bufio"
	"context"
	"io"
	"sync"
)

const (
	BufferedReaderSize = 32 * 1024
)

type bufferedReader struct {
	reader *bufio.Reader
	cancel context.CancelFunc
}

func (b *bufferedReader) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *bufferedReader) Reset(rd io.Reader, cancel context.CancelFunc) {
	b.reader.Reset(rd)
	b.cancel = cancel
}

var poolBufferedConn = sync.Pool{
	New: func() interface{} {
		return &bufferedReader{
			reader: bufio.NewReaderSize(nil, BufferedReaderSize),
		}
	},
}

func acquireBufferedReader(rd io.Reader, cancel context.CancelFunc) *bufferedReader {
	reader := poolBufferedConn.Get().(*bufferedReader)

	reader.Reset(rd, cancel)

	return reader
}

func releaseBufferedReader(conn *bufferedReader) {
	conn.Reset(nil, nil)
	poolBufferedConn.Put(conn)
}
