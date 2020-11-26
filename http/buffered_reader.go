package http

import (
	"bufio"
	"io"
	"sync"
)

const (
	BufferedReaderSize = 32 * 1024
)

type bufferedReader struct {
	reader   *bufio.Reader
	callback ResponseCallback
}

func (b *bufferedReader) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *bufferedReader) Reset(rd io.Reader, callback ResponseCallback) {
	b.reader.Reset(rd)
	b.callback = callback
}

var poolBufferedConn = sync.Pool{
	New: func() interface{} {
		return &bufferedReader{
			reader: bufio.NewReaderSize(nil, BufferedReaderSize),
		}
	},
}

func acquireBufferedReader(rd io.Reader, callback ResponseCallback) *bufferedReader {
	reader := poolBufferedConn.Get().(*bufferedReader)

	reader.Reset(rd, callback)

	return reader
}

func releaseBufferedReader(conn *bufferedReader) {
	conn.Reset(nil, nil)
	poolBufferedConn.Put(conn)
}
