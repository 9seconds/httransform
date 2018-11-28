package httransform

import "sync"

var layerStatePool sync.Pool

func init() {
	layerStatePool = sync.Pool{
		New: func() interface{} {
			return &LayerState{
				ctx: make(map[string]interface{}),
			}
		},
	}
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
