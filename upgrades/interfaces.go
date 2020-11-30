package upgrades

import (
	"context"
	"net"
)

type Interface interface {
	Manage(ctx context.Context, clientConn, netlocConn net.Conn)
}
