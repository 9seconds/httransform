package dialers

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

const connectResponseBufioBufferSize = 8192

var (
	poolConnectRequestByteBuffer = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	poolConnectResponseBufioReader = sync.Pool{
		New: func() interface {} {
			return bufio.NewReaderSize(nil, connectResponseBufioBufferSize)
		},
	}
)

func acquireConnectRequestByteBuffer() *bytes.Buffer {
	return poolConnectRequestByteBuffer.Get().(*bytes.Buffer)
}

func releaseConnectRequestByteBuffer(buf *bytes.Buffer) {
	buf.Reset()
	poolConnectRequestByteBuffer.Put(buf)
}

func acquireConnectResponseBufioReader(r io.Reader) *bufio.Reader {
	reader := poolConnectResponseBufioReader.Get().(*bufio.Reader)
	reader.Reset(r)

	return reader
}

func releaseConnectResponseBufioReader(reader *bufio.Reader) {
	reader.Reset(nil)
	poolConnectResponseBufioReader.Put(reader)
}
