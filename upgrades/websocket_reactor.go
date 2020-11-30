package upgrades

import "github.com/gobwas/ws/wsutil"

type WebsocketReactor interface {
	ClientMessage(wsutil.Message)
	ClientError(error)
	NetlocMessage(wsutil.Message)
	NetlocError(error)
}

type NoopWebsocketReactor struct{}

func (n NoopWebsocketReactor) ClientMessage(_ wsutil.Message) {}
func (n NoopWebsocketReactor) ClientError(_ error)            {}
func (n NoopWebsocketReactor) NetlocMessage(_ wsutil.Message) {}
func (n NoopWebsocketReactor) NetlocError(_ error)            {}
