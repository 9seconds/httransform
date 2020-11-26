package conns

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/9seconds/httransform/v2/events"
)

type TrafficConn struct {
	Parent  net.Conn
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

	n, err := t.Parent.Read(p)

	atomic.AddUint64(&t.readBytes, uint64(n))

	return n, err
}

func (t *TrafficConn) Write(p []byte) (int, error) {
	t.exclusiveMutex.RLock()
	defer t.exclusiveMutex.RUnlock()

	n, err := t.Parent.Write(p)

	atomic.AddUint64(&t.writtenBytes, uint64(n))

	return n, err
}

func (t *TrafficConn) Close() error {
	err := t.Parent.Close()

	t.closedOnce.Do(t.doClose)

	return err
}

func (t *TrafficConn) LocalAddr() net.Addr {
	return t.Parent.LocalAddr()
}

func (t *TrafficConn) RemoteAddr() net.Addr {
	return t.Parent.RemoteAddr()
}

func (t *TrafficConn) SetDeadline(tm time.Time) error {
	return t.Parent.SetDeadline(tm)
}

func (t *TrafficConn) SetReadDeadline(tm time.Time) error {
	return t.Parent.SetReadDeadline(tm)
}

func (t *TrafficConn) SetWriteDeadline(tm time.Time) error {
	return t.Parent.SetWriteDeadline(tm)
}

func (t *TrafficConn) doClose() {
	t.exclusiveMutex.Lock()
	defer t.exclusiveMutex.Unlock()

	meta := &events.TrafficMeta{
		ID:           t.ID,
		Addr:         t.Parent.RemoteAddr(),
		ReadBytes:    t.readBytes,
		WrittenBytes: t.writtenBytes,
	}

	t.Events.Send(t.Context, events.EventTypeTraffic, meta, t.ID)
}
