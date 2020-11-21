package layers

import (
	"fmt"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/PumpkinSeed/errors"
)

const (
	HeadersLayerKeyRequest  = "headers_layer__request"
	HeadersLayerKeyResponse = "headers_layer__response"
)

type StartHeadersLayer struct{}

func (s StartHeadersLayer) OnRequest(ctx *Context) error {
	requestHeaders := headers.AcquireHeaderSet()
	responseHeaders := headers.AcquireHeaderSet()
	headerWrapper := headers.NewRequestHeaderWrapper(&ctx.Request().Header)

	if err := requestHeaders.Init(headerWrapper); err != nil {
		return fmt.Errorf("cannot read request headers: %w", err)
	}

	ctx.Set(HeadersLayerKeyRequest, requestHeaders)
	ctx.Set(HeadersLayerKeyResponse, responseHeaders)

	return nil
}

func (s StartHeadersLayer) OnResponse(ctx *Context, err error) error {
	if requestHeadersRaw, ok := ctx.Get(HeadersLayerKeyRequest); ok {
		requestHeaders := requestHeadersRaw.(*headers.Headers)

		headers.ReleaseHeaderSet(requestHeaders)
		ctx.Delete(HeadersLayerKeyRequest)
	} else {
		panic("cannot find a request headers in the context")
	}

	if responseHeadersRaw, ok := ctx.Get(HeadersLayerKeyResponse); ok {
		responseHeaders := responseHeadersRaw.(*headers.Headers)

		if err2 := responseHeaders.Sync(); err2 != nil {
			return errors.Wrap(err, fmt.Errorf("cannot sync response headers: %w", err))
		}

		headers.ReleaseHeaderSet(responseHeaders)
		ctx.Delete(HeadersLayerKeyResponse)
	} else {
		panic("cannot find a request headers in the context")
	}

	return nil
}

type FinishHeadersLayer struct{}

func (f FinishHeadersLayer) OnRequest(ctx *Context) error {
	if requestHeadersRaw, ok := ctx.Get(HeadersLayerKeyRequest); ok {
		requestHeaders := requestHeadersRaw.(*headers.Headers)

		if err := requestHeaders.Sync(); err != nil {
			return fmt.Errorf("cannot sync request headers: %w", err)
		}
	} else {
		panic("cannot find a request headers")
	}

	return nil
}

func (f FinishHeadersLayer) OnResponse(ctx *Context, err error) error {
	if responseHeadersRaw, ok := ctx.Get(HeadersLayerKeyResponse); ok {
		responseHeaders := responseHeadersRaw.(*headers.Headers)
		responseHeadersWrapper := headers.NewResponseHeaderWrapper(&ctx.Response().Header)

		if err2 := responseHeaders.Init(responseHeadersWrapper); err2 != nil {
			return errors.Wrap(err, fmt.Errorf("cannot sync response headers: %w", err))
		}
	} else {
		panic("cannot find a response headers")
	}

	return nil
}
