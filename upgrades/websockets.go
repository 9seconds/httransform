package upgrades

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/gobwas/ws"
)

type WebsocketReactor interface {
	ClientMessage(*ws.Frame)
	ClientError(error)
	NetlocMessage(*ws.Frame)
	NetlocError(error)
}

type websocketUpgrader struct {
	reactor WebsocketReactor

	clientBuffer []byte
	netlocBuffer []byte
}

func (w *websocketUpgrader) Manage(ctx context.Context, clientConn, netlocConn net.Conn) {
	clientReader, clientWriter := io.Pipe()
	clientEnd := io.MultiWriter(clientConn, clientWriter)
	netlocReader, netlocWriter := io.Pipe()
	netlocEnd := io.MultiWriter(netlocConn, netlocWriter)
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-ctx.Done()
		clientConn.Close()
		netlocConn.Close()
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if frame, err := ws.ReadFrame(clientReader); err != nil {
				w.reactor.ClientError(err)
			} else {
				w.reactor.ClientMessage(&frame)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if frame, err := ws.ReadFrame(netlocReader); err != nil {
				w.reactor.NetlocError(err)
			} else {
				w.reactor.NetlocMessage(&frame)
			}
		}
	}()

	go w.manage(clientEnd, netlocConn, w.clientBuffer, cancel)

	w.manage(netlocEnd, clientConn, w.netlocBuffer, cancel)
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
