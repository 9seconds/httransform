package upgrades

import (
	"context"
	"net"
)

type Upgrader interface {
	Manage(ctx context.Context, clientConn, netlocConn net.Conn)
}
