package layers

import (
	"sync"

	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/headers"
)

var (
	poolLayerContext = sync.Pool{
		New: func() interface{} {
			return &LayerContext{}
		},
	}
)

func AcquireLayerContext(ctx *fasthttp.RequestCtx, isConnectMethod bool) *LayerContext {
	layerContext := poolLayerContext.Get().(*LayerContext)

	layerContext.RequestID = ctx.ID()
	layerContext.IsConnectMethod = isConnectMethod

	layerContext.RequestHeaders = headers.AcquireSet()
	layerContext.ResponseHeaders = headers.AcquireSet()
	layerContext.ConsumeRequest(&ctx.Request)

	layerContext.Request = &ctx.Request
	layerContext.Response = &ctx.Response

	layerContext.LocalAddr = ctx.Conn().LocalAddr()
	layerContext.RemoteAddr = ctx.Conn().RemoteAddr()

	layerContext.data = map[string]interface{}{}
	layerContext.ctx = ctx

	return layerContext
}

func ReleaseLayerContext(layerContext *LayerContext) {
	layerContext.Reset()
	poolLayerContext.Put(layerContext)
}
