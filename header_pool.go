package main

import "sync"

var (
	headerSetPool sync.Pool
	headersPool   sync.Pool
)

func init() {
	headerSetPool = sync.Pool{
		New: func() interface{} {
			return &HeaderSet{
				index:          map[string]int{},
				values:         []*Header{},
				removedHeaders: map[string]struct{}{},
			}
		},
	}
	headersPool = sync.Pool{
		New: func() interface{} {
			return &Header{}
		},
	}
}

func getHeader() *Header {
	header := headersPool.Get().(*Header)
	header.Key = header.Key[:0]
	header.Value = header.Value[:0]
	header.ID = ""

	return header
}

func getHeaderSet() *HeaderSet {
	set := headerSetPool.Get().(*HeaderSet)
	set.Clear()

	return set
}

func releaseHeaderSet(set *HeaderSet) {
	headerSetPool.Put(set)
}

func releaseHeader(header *Header) {
	headersPool.Put(header)
}
