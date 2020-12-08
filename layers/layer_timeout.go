package layers

import "time"

const TimeoutLayerKeyCancel = "timeout_layer__cancel"

type TimeoutLayer struct {
	Timeout time.Duration
}

func (t TimeoutLayer) OnRequest(ctx *Context) error {
	timer := time.AfterFunc(t.Timeout, ctx.Cancel)
	ctx.Set(TimeoutLayerKeyCancel, func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	})

	return nil
}

func (t TimeoutLayer) OnResponse(ctx *Context, err error) error {
	if cancel, ok := ctx.Get(TimeoutLayerKeyCancel); ok {
		cancel.(func())()
		ctx.Delete(TimeoutLayerKeyCancel)
	} else {
		panic("cannot find a cancel function in the context")
	}

	return err
}
