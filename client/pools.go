package client

import (
	"bufio"
	"sync"
)

const chunkedReaderBufferSize = 8 * 1024

var (
	poolBufferedReader = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, chunkedReaderBufferSize)
		},
	}
)
