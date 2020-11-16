package layers

import "time"

type TimeoutLayer struct {
	Timeout time.Duration
}

func (t TimeoutLayer) OnRequest(ctx *Context) error {
	timer := time.AfterFunc(t.Timeout, ctx.Cancel)
	ctx.Set("__timeout_layer_cancel__", timer.Stop)

	return nil
}

func (t TimeoutLayer) OnResponse(ctx *Context, err error) {
	if cancel, ok := ctx.Get("__timeout_layer_cancel__"); ok {
		cancel.(func() bool)()
	}
}
