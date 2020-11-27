package upgrades

import (
	"io"
	"net"
	"sync"
)

const (
	TCPUpgraderBufferSize = 10 * 1024
)

type tcpUpgrader struct {
	clientBuffer []byte
	netlocBuffer []byte
}

func (t *tcpUpgrader) Manage(clientConn, netlocConn net.Conn) {
	wg := &sync.WaitGroup{}

	wg.Add(2)

	go t.manage(clientConn, netlocConn, t.netlocBuffer, wg)

	go t.manage(netlocConn, clientConn, t.clientBuffer, wg)

	wg.Wait()
}

func (t *tcpUpgrader) manage(source io.ReadCloser, target io.WriteCloser, buf []byte, wg *sync.WaitGroup) {
	defer func() {
		source.Close()
		target.Close()
		wg.Done()
	}()

	io.CopyBuffer(target, source, buf)
}

func NewTCPUpgrader() Upgrader {
	return &tcpUpgrader{
		clientBuffer: make([]byte, TCPUpgraderBufferSize),
		netlocBuffer: make([]byte, TCPUpgraderBufferSize),
	}
}

var poolTCPUpgrader = sync.Pool{
	New: func() interface{} {
		return NewTCPUpgrader()
	},
}

func AcquireTCPUpgrader() Upgrader {
	return poolTCPUpgrader.Get().(Upgrader)
}

func ReleaseTCPUpgrader(up Upgrader) {
	poolTCPUpgrader.Put(up.(*tcpUpgrader))
}
