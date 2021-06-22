package upgrades

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	// WebsocketBufferSize defines a size of buffers required for
	// pumping data between peers. Websocket upgrader uses 2 buffers.
	WebsocketBufferSize = 10 * 1024
)

type websocketInterface struct {
	reactor WebsocketReactor

	clientBuffer []byte
	netlocBuffer []byte
}

func (w *websocketInterface) Manage(clientConn, netlocConn net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Add(2) // nolint: gomnd

	go func() {
		<-ctx.Done()
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

	go w.manage(clientWriter, netlocConn, w.clientBuffer, wg)

	go w.manage(netlocWriter, clientConn, w.netlocBuffer, wg)

	wg.Wait()
}

func (w *websocketInterface) consume(ctx context.Context,
	reader io.Reader,
	state ws.State,
	onMessage func(context.Context, wsutil.Message),
	onError func(context.Context, error)) {
	var err error

	messages := []wsutil.Message{}

	for {
		messages = messages[:0]

		select {
		case <-ctx.Done():
			return
		default:
			messages, err = wsutil.ReadMessage(reader, state, messages)

			for i := range messages {
				onMessage(ctx, messages[i])
			}

			if err != nil {
				onError(ctx, err)
			}
		}
	}
}

func (w *websocketInterface) manage(dst io.Writer,
	src io.ReadCloser,
	buf []byte,
	wg *sync.WaitGroup) {
	defer func() {
		src.Close()
		wg.Done()
	}()

	io.CopyBuffer(dst, src, buf) // nolint: errcheck
}

// NewWebsocket returns a new instance of Websocket upgrader.
//
// Websocket upgrader works in the same fashion as TCP upgrader: it
// pumps a data between 2 sockets. But at the same time, it reads
// messages and unmarshal them.
func NewWebsocket(reactor WebsocketReactor) Interface {
	return &websocketInterface{
		reactor:      reactor,
		clientBuffer: make([]byte, WebsocketBufferSize),
		netlocBuffer: make([]byte, WebsocketBufferSize),
	}
}

var poolWebsocket = sync.Pool{
	New: func() interface{} {
		return NewWebsocket(nil)
	},
}

// AcquireWebsocket returns a new Websocket upgrader from the pool.
func AcquireWebsocket(reactor WebsocketReactor) Interface {
	rv, _ := poolWebsocket.Get().(*websocketInterface)

	rv.reactor = reactor

	return rv
}

// ReleaseWebsocket returns Websocket back to the object pool.
func ReleaseWebsocket(up Interface) {
	value, _ := up.(*websocketInterface)
	value.reactor = nil

	poolWebsocket.Put(value)
}
