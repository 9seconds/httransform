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

type tcpWriterWrapper struct {
	ctx      context.Context
	callback func(context.Context, []byte)
}

func (t tcpWriterWrapper) Write(p []byte) (int, error) {
	t.callback(t.ctx, append([]byte(nil), p...))

	return len(p), nil
}

type tcpInterface struct {
	reactor      TCPReactor
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

	go t.manage(ctx,
		clientConn,
		netlocConn,
		t.reactor.NetlocBytes,
		t.reactor.NetlocError,
		t.netlocBuffer,
		cancel)

	t.manage(ctx,
		netlocConn,
		clientConn,
		t.reactor.ClientBytes,
		t.reactor.ClientError,
		t.clientBuffer,
		cancel)
}

func (t *tcpInterface) manage(ctx context.Context,
	src io.Reader,
	dst io.Writer,
	onWriteBytes func(context.Context, []byte),
	onWriteError func(context.Context, error),
	buf []byte,
	cancel context.CancelFunc) {
	defer cancel()

	writerWrapper := tcpWriterWrapper{
		ctx:      ctx,
		callback: onWriteBytes,
	}
	writer := io.MultiWriter(writerWrapper, dst)

	if _, err := io.CopyBuffer(writer, src, buf); err != nil {
		onWriteError(ctx, err)
	}
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

func AcquireTCP(reactor TCPReactor) Interface {
	rv := poolTCP.Get().(*tcpInterface)

	rv.reactor = reactor

	return rv
}

func ReleaseTCP(up Interface) {
	value := up.(*tcpInterface)
	value.reactor = nil

	poolTCP.Put(value)
}
