package dialers

import "errors"

// ErrNoIPs is returned if we tried to resolve DNS name and got no IPs.
var ErrNoIPs = errors.New("no ips")
