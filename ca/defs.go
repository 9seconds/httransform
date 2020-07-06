package ca

import (
	"time"

	"github.com/PumpkinSeed/errors"
)

const (
	DefaultMaxSize = 1024
	CertificateTTL = 24 * time.Hour

	WorkerCertificateTTL = CertificateTTL * 30
	RSAKeyLength         = 2048
)

var (
	ErrCA = errors.New("CA error")
	ErrContextClosed = errors.New("context is closed")

	ErrCAInvalidCertificates = errors.Wrap(errors.New("invalid ca certificate"), ErrCA)
)
