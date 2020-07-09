package layers

import (
	"net"
	"strconv"
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

func AcquireLayerContext() *LayerContext {
	layerContext := poolLayerContext.Get().(*LayerContext)

	layerContext.RequestHeaders = headers.AcquireSet()
	layerContext.ResponseHeaders = headers.AcquireSet()
	layerContext.data = map[string]interface{}{}

	return layerContext
}

func ReleaseLayerContext(layerContext *LayerContext) {
	layerContext.Reset()
	poolLayerContext.Put(layerContext)
}

func InitLayerContext(fasthttpCtx *fasthttp.RequestCtx, layerCtx *LayerContext) error {
	host, portStr, err := net.SplitHostPort(string(fasthttpCtx.Request.URI().Host()))
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	layerCtx.RequestID = fasthttpCtx.ID()
	layerCtx.ConnectHost = host
	layerCtx.ConnectPort = port

	layerCtx.Request = &fasthttpCtx.Request

	layerCtx.RequestHeaders.Reset()
	layerCtx.Request.Header.VisitAllInOrder(func(key, value []byte) {
		layerCtx.RequestHeaders.Add(string(key), string(value))
	})

	layerCtx.Response = &fasthttpCtx.Response

	layerCtx.ResponseHeaders.Reset()

	layerCtx.LocalAddr = fasthttpCtx.Conn().LocalAddr()
	layerCtx.RemoteAddr = fasthttpCtx.Conn().RemoteAddr()

	layerCtx.ctx = fasthttpCtx

	return nil
}
