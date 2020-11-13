package dialer

import "errors"

var (
	ErrNoIPs      = errors.New("no ips were resolved")
	ErrCannotDial = errors.New("failed to dial to any ")
)
