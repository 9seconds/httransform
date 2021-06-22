package headers

import (
	"bytes"
	"sync"
)

var poolBytesReader = sync.Pool{
	New: func() interface{} {
		return bytes.NewReader(nil)
	},
}

func acquireBytesReader(data []byte) *bytes.Reader {
	rd, _ := poolBytesReader.Get().(*bytes.Reader)

	rd.Reset(data)

	return rd
}

func releaseBytesReader(rd *bytes.Reader) {
	rd.Reset(nil)
	poolBytesReader.Put(rd)
}
