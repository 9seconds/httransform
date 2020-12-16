package upgrades

import (
	"context"
	"net"
)

// Interface is an interface for upgrader.
type Interface interface {
	// Manage is a blocking method which should be used to process both
	// ends of the connection. You have client and netloc connections
	// there. A method has to exit when contxt is cancelled.
	//
	// You do not need to close connections there. It is a
	// responsibility of httransform.
	Manage(ctx context.Context, clientConn, netlocConn net.Conn)
}
