package httransform

import (
	"bytes"
	"sync"
)

var (
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

	layerStatePool = sync.Pool{
		New: func() interface{} {
			return &LayerState{
				ctx: make(map[string]interface{}),
			}
		},
	}

	connectRequestBufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	dummyTracerPool = sync.Pool{
		New: func() interface{} {
			return &dummyTracer{}
		},
	}

	logTracerPool = sync.Pool{
		New: func() interface{} {
			return &logTracer{}
		},
	}
)

func getHeader() *Header {
	return headersPool.Get().(*Header)
}

func getHeaderSet() *HeaderSet {
	return headerSetPool.Get().(*HeaderSet)
}

func getLayerState() *LayerState {
	return layerStatePool.Get().(*LayerState)
}

func releaseHeaderSet(set *HeaderSet) {
	set.Clear()
	headerSetPool.Put(set)
}

func releaseHeader(header *Header) {
	header.Key = header.Key[:0]
	header.Value = header.Value[:0]
	header.ID = ""
	headersPool.Put(header)
}

func releaseLayerState(state *LayerState) {
	for k := range state.ctx {
		delete(state.ctx, k)
	}
	layerStatePool.Put(state)
}
