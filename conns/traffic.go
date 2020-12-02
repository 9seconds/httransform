package conns

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/9seconds/httransform/v2/events"
)

// TrafficConn defines a wrapper around net.Conn which calculates
// a traffic which was passed via this connection and reports on
// connection closing.
//
// It calculates a number of bytes which were read and written
// (separately) and reports to a given events.EventChannel with
// events.TrafficMeta instance.
//
// A closing is done once. So, you can make as much consequentive calls
// as you want. You are going to be reported only on a first one.
type TrafficConn struct {
	// Actual connection this wrapper is wrapping.
	net.Conn

	// A context for this connection. This is used only for sending
	// metadata on connection close. If context is done and connection
	// is closed, then TrafficMeta won't be send.
	Context context.Context

	// ID of the connection. If this class is used by httransform
	// itself. then ID is identifier of RequestID field from
	// layers.Context. But if you use it from your own library,
	// you can put any identifier you like.
	ID string

	// A channel for httransform event stream.
	Events *events.Channel

	readBytes      uint64
	writtenBytes   uint64
	closedOnce     sync.Once
	exclusiveMutex sync.RWMutex
}

// Read requires to conform io.ReadWriteCloser interface.
func (t *TrafficConn) Read(p []byte) (int, error) {
	t.exclusiveMutex.RLock()
	defer t.exclusiveMutex.RUnlock()

	n, err := t.Conn.Read(p)

	atomic.AddUint64(&t.readBytes, uint64(n))

	return n, err // nolint: wrapcheck
}

// Write requires to conform io.ReadWriteCloser interface.
func (t *TrafficConn) Write(p []byte) (int, error) {
	t.exclusiveMutex.RLock()
	defer t.exclusiveMutex.RUnlock()

	n, err := t.Conn.Write(p)

	atomic.AddUint64(&t.writtenBytes, uint64(n))

	return n, err // nolint: wrapcheck
}

// Close requires to conform io.ReadWriteCloser interface.
func (t *TrafficConn) Close() error {
	err := t.Conn.Close()

	t.closedOnce.Do(t.doClose)

	return err // nolint: wrapcheck
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
