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

type tcpInterface struct {
	clientBuffer []byte
	netlocBuffer []byte
}

func (t *tcpInterface) Manage(ctx context.Context, clientConn, netlocConn net.Conn) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-ctx.Done()
		clientConn.Close()
		netlocConn.Close()
	}()

	go t.manage(clientConn, netlocConn, t.netlocBuffer, cancel)

	t.manage(netlocConn, clientConn, t.clientBuffer, cancel)
}

func (t *tcpInterface) manage(src io.Reader, dst io.Writer, buf []byte, cancel context.CancelFunc) {
	defer cancel()

	io.CopyBuffer(dst, src, buf) // nolint: errcheck
}

func NewTCP() Interface {
	return &tcpInterface{
		clientBuffer: make([]byte, TCPBufferSize),
		netlocBuffer: make([]byte, TCPBufferSize),
	}
}

var poolTCP = sync.Pool{
	New: func() interface{} {
		return NewTCP()
	},
}

func AcquireTCP() Interface {
	return poolTCP.Get().(Interface)
}

func ReleaseTCP(up Interface) {
	poolTCP.Put(up.(*tcpInterface))
}
