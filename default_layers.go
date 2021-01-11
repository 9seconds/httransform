package httransform

import (
	"github.com/9seconds/httransform/v2/errors"
	"github.com/9seconds/httransform/v2/layers"
)

type layerStartHeaders struct{}

func (l layerStartHeaders) OnRequest(ctx *layers.Context) error {
	return nil
}

func (l layerStartHeaders) OnResponse(ctx *layers.Context, err error) error {
	if err2 := ctx.ResponseHeaders.Push(); err2 != nil {
		return &errors.Error{
			Message: "cannot sync response headers",
			Err:     err,
		}
	}

	return err
}

type layerFinishHeaders struct{}

func (l layerFinishHeaders) OnRequest(ctx *layers.Context) error {
	if err := ctx.RequestHeaders.Push(); err != nil {
		return &errors.Error{
			Message: "cannot sync request headers",
			Err:     err,
		}
	}

	return nil
}

func (l layerFinishHeaders) OnResponse(ctx *layers.Context, err error) error {
	if err2 := ctx.ResponseHeaders.Pull(); err2 != nil {
		return &errors.Error{
			Message: "cannot read response headers",
			Err:     err,
		}
	}

	return err
}
