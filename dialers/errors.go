package dialers

import "github.com/9seconds/httransform/v2/errors"

// ErrNoIPs is returned if we tried to resolve DNS name and got no IPs.
var ErrNoIPs = &errors.Error{
	Message: "no ips",
	Code:    "dns_no_ips",
}
