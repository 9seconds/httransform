package dialers

import (
	"time"

	"github.com/PumpkinSeed/errors"
)

const (
	DNSRefreshEvery = time.Minute

	SimpleDialerTimeout = 30 * time.Second

	PooledDialerTimeout    = 30 * time.Second
	PooledDialerGCEvery    = time.Second
	PooledDialerStaleAfter = time.Minute

	ConnectionPoolGCEvery    = time.Second
	ConnectionPoolStaleAfter = 30 * time.Second
)

var (
	ErrDialer              = errors.New("dialer error")
	ErrCannotSplitHostPort = errors.Wrap(errors.New("cannot split host/port"), ErrDialer)
	ErrDNSError            = errors.Wrap(errors.New("dns failure"), ErrDialer)
	ErrNoIPs               = errors.Wrap(errors.New("no ips were found"), ErrDialer)
	ErrCannotDial          = errors.Wrap(errors.New("cannot dial"), ErrDialer)
	ErrContextClosed       = errors.New("context is closed")
)
