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
)

func MakeDefaultExecutor(ctx context.Context, dialTimeout time.Duration) Executor {
	if dialTimeout == 0 {
		dialTimeout = DefaultDialTimeout
	}

	client := client.NewClient(dialer.NewBase(ctx, dialTimeout), false)

	return func(ctx *layers.Context) error {
		return client.Do(ctx, ctx.Request(), ctx.Response())
	}
}
