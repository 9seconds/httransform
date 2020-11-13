package executor

import (
	"context"

	"github.com/9seconds/httransform/v2/client"
	"github.com/9seconds/httransform/v2/layers"
)

type Executor func(*layers.Context)

func MakeDefaultExecutor(ctx context.Context) Executor {
	httpClient := client.NewClient(ctx)

	return func(ctx *layers.Context) {
		if err := httpClient.Do(ctx, ctx.Request(), ctx.Response()); err != nil {
			ctx.Error(err)
		}
	}
}
