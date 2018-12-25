package client

import (
	"bufio"
	"sync"
)

const chunkedReaderBufferSize = 16 * 1024

var (
	poolBufferedReader = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, chunkedReaderBufferSize)
		},
	}
)
