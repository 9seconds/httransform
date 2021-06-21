package upgrades

import (
	"context"
	"io"
	"net"
	"sync"
)

const (
	// TCPBufferSize defines a size of the buffer for TCP upgrader.
	// Actually, TCP upgrader allocates 2 buffers of this size.
	TCPBufferSize = 10 * 1024
)

type tcpWriterWrapper struct {
	ctx      context.Context
	callback func(context.Context, []byte)
	buf      []byte
}

func (t *tcpWriterWrapper) Write(p []byte) (int, error) {
	t.buf = append(t.buf, p...)

	t.callback(t.ctx, t.buf)

	t.buf = t.buf[:0]

	return len(p), nil
}

type tcpInterface struct {
	reactor      TCPReactor
	clientBuffer []byte
	netlocBuffer []byte
}

func (t *tcpInterface) Manage(clientConn, netlocConn net.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Add(2) // nolint: gomnd

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
		wg)

	go t.manage(ctx,
		netlocConn,
		clientConn,
		t.reactor.ClientBytes,
		t.reactor.ClientError,
		t.clientBuffer,
		wg)

	wg.Wait()
}

func (t *tcpInterface) manage(ctx context.Context,
	src io.ReadCloser,
	dst io.Writer,
	onWriteBytes func(context.Context, []byte),
	onWriteError func(context.Context, error),
	buf []byte,
	wg *sync.WaitGroup) {
	defer func() {
		src.Close()
		wg.Done()
	}()

	writerWrapper := &tcpWriterWrapper{
		ctx:      ctx,
		callback: onWriteBytes,
	}
	writer := io.MultiWriter(writerWrapper, dst)

	if _, err := io.CopyBuffer(writer, src, buf); err != nil {
		onWriteError(ctx, err)
	}
}

// NewTCP returns a new instance of TCP upgrader.
//
// TCP upgrader is really simple and dumb: it pump data from one socket
// to another. Given reactor just looks at what is happening there.
// Another important consideration is that reactor cannot change a data.
// It always gets a copy.
func NewTCP(reactor TCPReactor) Interface {
	return &tcpInterface{
		clientBuffer: make([]byte, TCPBufferSize),
		netlocBuffer: make([]byte, TCPBufferSize),
		reactor:      reactor,
	}
}

var poolTCP = sync.Pool{
	New: func() interface{} {
		return NewTCP(nil)
	},
}

// AcquireTCP returns a new TCP upgrader from the pool.
func AcquireTCP(reactor TCPReactor) Interface {
	rv, _ := poolTCP.Get().(*tcpInterface)

	rv.reactor = reactor

	return rv
}

// ReleaseTCP returns TCP upgrader back to the object pool.
func ReleaseTCP(up Interface) {
	value, _ := up.(*tcpInterface)
	value.reactor = nil

	poolTCP.Put(value)
}
