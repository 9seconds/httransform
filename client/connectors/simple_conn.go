package connectors

import "net"

type simpleConn struct {
	net.TCPConn
}

func (s *simpleConn) Release() {
	s.Close()
}
