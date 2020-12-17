package upgrades

import "net"

// Interface is an interface for upgrader.
type Interface interface {
	// Manage is a blocking method which should be used to process both
	// ends of the connection. You have client and netloc connections
	// there. You do not need to close them there. It is a responsibility
	// of httransform.
	Manage(clientConn, netlocConn net.Conn)
}
