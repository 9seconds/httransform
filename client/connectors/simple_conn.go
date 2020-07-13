package connectors

import "net"

type simpleConn struct {
	net.Conn
}

func (s *simpleConn) Release() {
	s.Close()
}
