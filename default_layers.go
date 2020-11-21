package main

import (
	"fmt"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/PumpkinSeed/errors"
)

type layerStartHeaders struct{}

func (l layerStartHeaders) OnRequest(ctx *layers.Context) error {
	requestHeaders := headers.NewRequestHeaderWrapper(&ctx.Request().Header)

	if err := ctx.RequestHeaders.Init(requestHeaders); err != nil {
		return fmt.Errorf("cannot read request headers: %w", err)
	}

	return nil
}

func (l layerStartHeaders) OnResponse(ctx *layers.Context, err error) error {
	if err2 := ctx.ResponseHeaders.Sync(); err2 != nil {
		return errors.Wrap(err, fmt.Errorf("cannot sync response headers: %w", err))
	}

	return err
}

type layerFinishHeaders struct{}

func (l layerFinishHeaders) OnRequest(ctx *layers.Context) error {
	if err := ctx.RequestHeaders.Sync(); err != nil {
		return fmt.Errorf("cannot sync request headers: %w", err)
	}

	return nil
}

func (l layerFinishHeaders) OnResponse(ctx *layers.Context, err error) error {
	responseHeaders := headers.NewResponseHeaderWrapper(&ctx.Response().Header)

	if err2 := ctx.ResponseHeaders.Init(responseHeaders); err2 != nil {
		return errors.Wrap(err, fmt.Errorf("cannot read response headers: %w", err))
	}

	return err
}
