package httransform

import (
	"bytes"
	"sync"
)

var (
	headerSetPool            sync.Pool
	headersPool              sync.Pool
	layerStatePool           sync.Pool
	connectRequestBufferPool sync.Pool
)

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

func getLayerState() *LayerState {
	state := layerStatePool.Get().(*LayerState)
	for k := range state.ctx {
		delete(state.ctx, k)
	}

	return state
}

func releaseLayerState(state *LayerState) {
	layerStatePool.Put(state)
}

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
}
