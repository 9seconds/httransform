package executor

import (
	"github.com/9seconds/httransform/v2/client"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/layers"
)

func MakeDefaultExecutor(dialer dialers.Dialer) Executor {
	client := client.NewClient(dialer)

	return func(ctx *layers.Context) error {
		return client.Do(ctx, ctx.Request(), ctx.Response())
	}
}
