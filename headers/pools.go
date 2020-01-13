package headers

import "sync"

var (
	poolHeader = sync.Pool{
		New: func() interface{} {
			return &Header{}
		},
	}
	poolSet = sync.Pool{
		New: func() interface{} {
			return &Set{
				index:          map[string]int{},
				data:           []*Header{},
				removedHeaders: map[string]struct{}{},
			}
		},
	}
)

func AcquireHeader() *Header {
	return poolHeader.Get().(*Header)
}

func ReleaseHeader(header *Header) {
	if header != nil {
		header.Reset()
		poolHeader.Put(header)
	}
}

func AcquireSet() *Set {
	return poolSet.Get().(*Set)
}

func ReleaseSet(set *Set) {
	if set != nil {
		set.Reset()
		poolSet.Put(set)
	}
}
