package dialer

import (
	"net"
	"time"
)

type BaseDialer func(string, time.Duration) (net.Conn, error)
