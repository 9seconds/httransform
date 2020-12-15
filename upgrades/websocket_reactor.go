package upgrades

import (
	"context"

	"github.com/gobwas/ws/wsutil"
)

// WebsocketReactor defines a set of callbacks user has to use
// to process a websocket message.
//
// These callbacks are executed in blocking mode. Please do necessary
// things to prevent corks and bottlenecks there.
type WebsocketReactor interface {
	// ClientMessage is executed when netloc sends a message to a client.
	ClientMessage(context.Context, wsutil.Message)

	// ClientError is executed when netloc failed to send a message to a
	// client or message could not be processed.
	ClientError(context.Context, error)

	// NetlocMessage is executed when client sends a message to a netloc.
	NetlocMessage(context.Context, wsutil.Message)

	// NetlocError is executed when client failed to send a message to a
	// netloc or message could not be processed.
	NetlocError(context.Context, error)
}

// NoopWebsocketReactor is Websocket reactor which does nothing.
type NoopWebsocketReactor struct{}

// ClientMessage conforms WebsocketReactor interface.
func (n NoopWebsocketReactor) ClientMessage(_ context.Context, _ wsutil.Message) {}

// ClientError conforms WebsocketReactor interface.
func (n NoopWebsocketReactor) ClientError(_ context.Context, _ error) {}

// NetlocMessage conforms WebsocketReactor interface.
func (n NoopWebsocketReactor) NetlocMessage(_ context.Context, _ wsutil.Message) {}

// NetlocError conforms WebsocketReactor interface.
func (n NoopWebsocketReactor) NetlocError(_ context.Context, _ error) {}
