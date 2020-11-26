package conns

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/9seconds/httransform/v2/events"
)

type TrafficConn struct {
	net.Conn

	Context context.Context
	ID      string
	Events  events.EventChannel

	readBytes      uint64
	writtenBytes   uint64
	closedOnce     sync.Once
	exclusiveMutex sync.RWMutex
}

func (t *TrafficConn) Read(p []byte) (int, error) {
	t.exclusiveMutex.RLock()
	defer t.exclusiveMutex.RUnlock()

	n, err := t.Conn.Read(p)

	atomic.AddUint64(&t.readBytes, uint64(n))

	return n, err
}

func (t *TrafficConn) Write(p []byte) (int, error) {
	t.exclusiveMutex.RLock()
	defer t.exclusiveMutex.RUnlock()

	n, err := t.Conn.Write(p)

	atomic.AddUint64(&t.writtenBytes, uint64(n))

	return n, err
}

func (t *TrafficConn) Close() error {
	err := t.Conn.Close()

	t.closedOnce.Do(t.doClose)

	return err
}

func (t *TrafficConn) doClose() {
	t.exclusiveMutex.Lock()
	defer t.exclusiveMutex.Unlock()

	meta := &events.TrafficMeta{
		ID:           t.ID,
		Addr:         t.RemoteAddr(),
		ReadBytes:    t.readBytes,
		WrittenBytes: t.writtenBytes,
	}

	t.Events.Send(t.Context, events.EventTypeTraffic, meta, t.ID)
}
