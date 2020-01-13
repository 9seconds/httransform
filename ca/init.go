package ca

import (
	"crypto/tls"
	"time"
)

const (
	DefaultMaxSize = 1024

	RSAKeyLength = 2048
)

var (
	DefaultTLSConfig = &tls.Config{
		InsecureSkipVerify: true, // nolint: gosec
	}
)

var bigBangTime = time.Unix(0, 0)
