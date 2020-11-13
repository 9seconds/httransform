package client

import "errors"

var (
	ErrNoIPs = errors.New("no ips were resolved")
)
