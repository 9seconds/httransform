package upgrades

import (
	"context"
	"io"
	"net"
	"sync"
)

const (
	TCPBufferSize = 10 * 1024
)

type tcpUpgrader struct {
	clientBuffer []byte
	netlocBuffer []byte
}

func (t *tcpUpgrader) Manage(ctx context.Context, clientConn, netlocConn net.Conn) {
    ctx, cancel := context.WithCancel(ctx)

    go func() {
        <-ctx.Done()
        clientConn.Close()
        netlocConn.Close()
    }()

    go t.manage(clientConn, netlocConn, t.netlocBuffer, cancel)

    t.manage(netlocConn, clientConn, t.clientBuffer, cancel)
}

func (t *tcpUpgrader) manage(src io.Reader, dst io.Writer, buf []byte, cancel context.CancelFunc) {
    defer cancel()

	io.CopyBuffer(dst, src, buf) // nolint: errcheck
}

func NewTCP() Upgrader {
	return &tcpUpgrader{
		clientBuffer: make([]byte, TCPBufferSize),
		netlocBuffer: make([]byte, TCPBufferSize),
	}
}

var poolTCP = sync.Pool{
	New: func() interface{} {
		return NewTCP()
	},
}

func AcquireTCP() Upgrader {
	return poolTCP.Get().(Upgrader)
}

func ReleaseTCP(up Upgrader) {
	poolTCP.Put(up.(*tcpUpgrader))
}
