package client3

import (
	"time"

	"github.com/PumpkinSeed/errors"
)

const (
	DefaultHTTPTimeout = 3 * time.Minute
	DefaultHTTPPort    = "80"
	DefaultHTTPSPort   = "443"
)

var (
	ErrClient = errors.New("http client error")
)
