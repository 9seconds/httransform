package upgrades

import "net"

type Upgrader interface {
    Manage(clientConn, netlocConn net.Conn)
}
