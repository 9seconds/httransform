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

func (n NoopWebsocketReactor) ClientMessage(_ wsutil.Message) {}
func (n NoopWebsocketReactor) ClientError(_ error)            {}
func (n NoopWebsocketReactor) NetlocMessage(_ wsutil.Message) {}
func (n NoopWebsocketReactor) NetlocError(_ error)            {}
