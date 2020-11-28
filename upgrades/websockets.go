package upgrades

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebsocketReactor interface {
	ClientMessage(wsutil.Message)
	ClientError(error)
	NetlocMessage(wsutil.Message)
	NetlocError(error)
}

type websocketUpgrader struct {
	reactor WebsocketReactor

	clientBuffer []byte
	netlocBuffer []byte
}

func (w *websocketUpgrader) Manage(ctx context.Context, clientConn, netlocConn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-ctx.Done()
		cancel()
		clientConn.Close()
		netlocConn.Close()
	}()

	clientPipeReader, clientPipeWriter := io.Pipe()
	clientWriter := io.MultiWriter(clientConn, clientPipeWriter)

	go w.consume(ctx,
		clientPipeReader,
		ws.StateClientSide,
		w.reactor.ClientMessage,
		w.reactor.ClientError)

	netlocPipeReader, netlocPipeWriter := io.Pipe()
	netlocWriter := io.MultiWriter(netlocConn, netlocPipeWriter)

	go w.consume(ctx,
		netlocPipeReader,
		ws.StateServerSide,
		w.reactor.NetlocMessage,
		w.reactor.NetlocError)

	go w.manage(clientWriter, netlocConn, w.clientBuffer, cancel)

	w.manage(netlocWriter, clientConn, w.netlocBuffer, cancel)
}

func (w *websocketUpgrader) consume(ctx context.Context,
	reader io.Reader,
	state ws.State,
	onMessage func(wsutil.Message),
	onError func(error)) {
	var err error

	messages := []wsutil.Message{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			messages, err = wsutil.ReadMessage(reader, state, messages[:0])

			for _, v := range messages {
				onMessage(v)
			}

			if err != nil {
				onError(err)
			}
		}
	}
}

func (w *websocketUpgrader) manage(dst io.Writer, src io.Reader, buf []byte, cancel context.CancelFunc) {
	defer cancel()

	io.CopyBuffer(dst, src, buf)
}

func NewWebsocket(reactor WebsocketReactor) Upgrader {
	return &websocketUpgrader{
		reactor:      reactor,
		clientBuffer: make([]byte, TCPBufferSize),
		netlocBuffer: make([]byte, TCPBufferSize),
	}
}

var poolWebsocket = sync.Pool{
	New: func() interface{} {
		return NewWebsocket(nil)
	},
}

func AcquireWebsocket(reactor WebsocketReactor) Upgrader {
	rv := poolWebsocket.Get().(*websocketUpgrader)

	rv.reactor = reactor

	return rv
}

func ReleaseWebsocket(up Upgrader) {
	up_ := up.(*websocketUpgrader)
	up_.reactor = nil

	poolWebsocket.Put(up_)
}
