package conns

import (
	"net"

	"github.com/valyala/fasthttp"
)

type hijackedConnInterface interface {
	UnsafeConn() net.Conn
}

// FixedHijackHandler defines a hijack handler which should be
// used along with FixHijackHandler.
type FixedHijackHandler func(net.Conn) bool

// FixHijackHandler fixes some inconveninces of
// fasthttp.HijackHandler.
//
// 1. It passes a real connection, not fasthttp.hijackConn
//
// 2. Handler returns a boolean value which defines if connection
// should be closed or not.
//
// A reason for that is fasthttp.Server which pools such connections.
// And returns connection back to pool on calling Close(). So, double
// closing is the same as double free: a segmentation fault. Usually
// users do not struggle but for our usecases we actually have to do
// double hijacking (TLS tunneling + websockets) so double closing is a
// real issue.
func FixHijackHandler(callback FixedHijackHandler) fasthttp.HijackHandler {
	return func(conn net.Conn) {
		for {
			if unwrapped, ok := conn.(hijackedConnInterface); ok {
				conn = unwrapped.UnsafeConn()
			} else {
				break
			}
		}

		if callback(conn) {
			conn.Close()
		}
	}
}
