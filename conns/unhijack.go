package conns

import (
	"net"

	"github.com/valyala/fasthttp"
)

type hijackedConnInterface interface {
	UnsafeConn() net.Conn
}

func FixedHijackHandler(callback func(net.Conn) bool) fasthttp.HijackHandler {
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
