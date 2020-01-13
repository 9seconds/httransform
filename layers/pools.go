package layers

import (
	"sync"

	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/headers"
)

var (
	poolLayerContext = sync.Pool{
		New: func() interface{} {
			return &LayerContext{
				data: map[string]interface{}{},
			}
		},
	}
)

func AcquireLayerContext(ctx *fasthttp.RequestCtx) *LayerContext {
	lc := poolLayerContext.Get().(*LayerContext)

	lc.Tunneled = ctx.IsConnect()
	lc.RequestHeaders = headers.AcquireSet()
	lc.ResponseHeaders = headers.AcquireSet()
	lc.ctx = ctx
	lc.RequestID = ctx.ID()
	lc.Request = &ctx.Request
	lc.Response = &ctx.Response
	lc.LocalAddr = ctx.LocalAddr()
	lc.RemoteAddr = ctx.RemoteAddr()

	return lc
}

func ReleaseLayerContext(lc *LayerContext) {
	if lc != nil {
		lc.Reset()
		poolLayerContext.Put(lc)
	}
}
