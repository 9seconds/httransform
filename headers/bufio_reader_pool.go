package headers

import (
	"bufio"
	"io"
	"sync"
)

const bufioReaderSize = 1024 * 64

var poolBufioReader = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderSize(nil, bufioReaderSize)
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
