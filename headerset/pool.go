package headerset

import "sync"

var headersPool sync.Pool

func init() {
	headersPool = sync.Pool{
		New: func() interface{} {
			return &HeaderSet{
				index:          map[string]int{},
				values:         []*Header{},
				removedHeaders: map[string]struct{}{},
			}
		},
	}
}

func GetHeaderSet() *HeaderSet {
	set := headersPool.Get().(*HeaderSet)
	set.Clear()

	return set
}

func ReleaseHeaderSet(set *HeaderSet) {
	headersPool.Put(set)
}
