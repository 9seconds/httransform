package layers

import "time"

// TimeoutLayerKeyCancel defines a key which is used in context to store
// some internal data.
const TimeoutLayerKeyCancel = "timeout_layer__cancel"

// TimeoutLayer defines a layer which control a time of execution of the
// request. So, if you want to limit an execution time by 1 minute, this
// is a right place to do.
//
// The only flaw is that IF we stopped to process a response and started
// to pump it back, this middleware does not control that.
type TimeoutLayer struct {
	// Timeout defines a hard time limit of the request.
	Timeout time.Duration
}

// OnRequest conforms Layer interface.
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

// OnResponse conforms Layer interface.
func (t TimeoutLayer) OnResponse(ctx *Context, err error) error {
	if cancel, ok := ctx.Get(TimeoutLayerKeyCancel); ok {
		cancel.(func())()
		ctx.Delete(TimeoutLayerKeyCancel)
	} else {
		panic("cannot find a cancel function in the context")
	}

	return err
}
