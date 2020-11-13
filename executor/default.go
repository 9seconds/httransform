package executor

import (
	"context"
	"time"

	"github.com/9seconds/httransform/v2/client"
	"github.com/9seconds/httransform/v2/dialer"
	"github.com/9seconds/httransform/v2/layers"
)

const (
	DefaultDialTimeout = 20 * time.Second
	DefaultHTTPTimeout = 3 * time.Minute
)

func MakeDefaultExecutor(ctx context.Context, dialTimeout, httpTimeout time.Duration) Executor {
	if dialTimeout == 0 {
		dialTimeout = DefaultDialTimeout
	}

	if httpTimeout == 0 {
		httpTimeout = DefaultHTTPTimeout
	}

	client := client.NewClient(dialer.NewBase(ctx, dialTimeout), httpTimeout, false)

	return func(ctx *layers.Context) {
		if err := client.Do(ctx, ctx.Request(), ctx.Response()); err != nil {
			ctx.Error(err)
		}
	}
}
