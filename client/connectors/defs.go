package connectors

import (
	"time"

	"github.com/PumpkinSeed/errors"
)

const (
	DNSRefreshEvery = time.Minute

	SimpleConnectorTimeout = 30 * time.Second

	PooledConnectorTimeout    = 30 * time.Second
	PooledConnectorGCEvery    = time.Second
	PooledConnectorStaleAfter = time.Minute

	PooledConnectorConnectionPoolGCEvery    = time.Second
	PooledConnectorConnectionPoolStaleAfter = 30 * time.Second
)

var (
	ErrConnector           = errors.New("dialer error")
	ErrCannotSplitHostPort = errors.Wrap(errors.New("cannot split host/port"), ErrConnector)
	ErrDNSError            = errors.Wrap(errors.New("dns failure"), ErrConnector)
	ErrNoIPs               = errors.Wrap(errors.New("no ips were found"), ErrConnector)
	ErrCannotDial          = errors.Wrap(errors.New("cannot dial"), ErrConnector)
	ErrContextClosed       = errors.New("context is closed")
)
