package http

import (
	"bufio"
	"io"
	"sync"
)

const (
	// BufferedReaderSize defines a size of buffer for reading from the
	// socket.
	BufferedReaderSize = 32 * 1024
)

var poolBufioReader = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderSize(nil, BufferedReaderSize)
	},
}

func acquireBufioReader(rd io.Reader) *bufio.Reader {
	reader, _ := poolBufioReader.Get().(*bufio.Reader)

	reader.Reset(rd)

	return reader
}

func releaseBufioReader(rd *bufio.Reader) {
	rd.Reset(nil)
	poolBufioReader.Put(rd)
}
