package headers

import (
	"strings"
	"sync"
)

var (
	poolHeader = sync.Pool{
		New: func() interface{} {
			return &Header{}
		},
	}
	poolSet = sync.Pool{
		New: func() interface{} {
			return &Set{}
		},
	}
)

func AcquireHeader(name, value string) *Header {
	header := poolHeader.Get().(*Header)
	header.id = strings.ToLower(name)
	header.Name = name
	header.Value = value

	return header
}

func AcquireSet() *Set {
	return poolSet.Get().(*Set)
}

func ReleaseHeader(header *Header) {
	header.Reset()
	poolHeader.Put(header)
}

func ReleaseSet(s *Set) {
	s.Reset()
	poolSet.Put(s)
}
