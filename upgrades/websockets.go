package upgrades

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type websocketInterface struct {
	reactor WebsocketReactor

	clientBuffer []byte
	netlocBuffer []byte
}

func (w *websocketInterface) Manage(ctx context.Context, clientConn, netlocConn net.Conn) {
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

func (w *websocketInterface) consume(ctx context.Context,
	reader io.Reader,
	state ws.State,
	onMessage func(context.Context, wsutil.Message),
	onError func(context.Context, error)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			messages, err := wsutil.ReadMessage(reader, state, nil)

			for _, v := range messages {
				onMessage(ctx, v)
			}

			if err != nil {
				onError(ctx, err)
			}
		}
	}
}

func (w *websocketInterface) manage(dst io.Writer, src io.Reader, buf []byte, cancel context.CancelFunc) {
	defer cancel()

	io.CopyBuffer(dst, src, buf) // nolint: errcheck
}

func NewWebsocket(reactor WebsocketReactor) Interface {
	return &websocketInterface{
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

func AcquireWebsocket(reactor WebsocketReactor) Interface {
	rv := poolWebsocket.Get().(*websocketInterface)

	rv.reactor = reactor

	return rv
}

func ReleaseWebsocket(up Interface) {
	value := up.(*websocketInterface)
	value.reactor = nil

	poolWebsocket.Put(value)
}
