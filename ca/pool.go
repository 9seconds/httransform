package ca

import (
	"sync"

	"github.com/karlseguin/ccache"
)

var (
	signRequestPool  sync.Pool
	signResponsePool sync.Pool
)

type signResponse struct {
	item ccache.TrackedItem
	err  error
}

type signRequest struct {
	host         string
	responseChan chan *signResponse
}

func init() {
	signRequestPool = sync.Pool{
		New: func() interface{} {
			return &signRequest{
				responseChan: make(chan *signResponse),
			}
		},
	}
	signRequestPool = sync.Pool{
		New: func() interface{} {
			return &signResponse{}
		},
	}
}
