package main

import "sync"

var layerStatePool sync.Pool

func init() {
	layerStatePool = sync.Pool{
		New: func() interface{} {
			return &LayerState{}
		},
	}
}

func getLayerState() *LayerState {
	return layerStatePool.Get().(*LayerState)
}

func releaseLayerState(statek *LayerState) {
	layerStatePool.Put(statek)
}
