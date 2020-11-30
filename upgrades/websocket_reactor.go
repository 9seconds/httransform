package upgrades

import (
	"context"

	"github.com/gobwas/ws/wsutil"
)

type WebsocketReactor interface {
	ClientMessage(context.Context, wsutil.Message)
	ClientError(context.Context, error)
	NetlocMessage(context.Context, wsutil.Message)
	NetlocError(context.Context, error)
}

type NoopWebsocketReactor struct{}

func (n NoopWebsocketReactor) ClientMessage(_ context.Context, _ wsutil.Message) {}
func (n NoopWebsocketReactor) ClientError(_ context.Context, _ error)            {}
func (n NoopWebsocketReactor) NetlocMessage(_ context.Context, _ wsutil.Message) {}
func (n NoopWebsocketReactor) NetlocError(_ context.Context, _ error)            {}
