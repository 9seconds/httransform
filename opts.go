package httransform

import (
	"time"

	"github.com/valyala/fasthttp"
)

// Defaults for ServerOpts.
const (
	// DefaultConcurrency is the number of concurrent connections for the
	// service.
	DefaultConcurrency = fasthttp.DefaultConcurrency

	// DefaultReadBufferSize is default size of read buffer for the
	// connection in bytes.
	DefaultReadBufferSize = 4096

	// DefaultWriteBufferSize is default size of write buffer for the
	// connection in bytes.
	DefaultWriteBufferSize = 4096

	// DefaultReadTimeout is the default timeout httransform uses to read
	// from user request.
	DefaultReadTimeout = time.Minute

	// DefaultWriteTimeout is the default timeout httransform uses to write
	// to user request.
	DefaultWriteTimeout = time.Minute

	// DefaultTLSCertCacheSize is the number of items we store in LRU cache
	// of generated TLS certificates before starting to prune obsoletes.
	DefaultTLSCertCacheSize = 10000

	// DefaultTLSCertCachePrune is the number of items we prune from LRU
	// cache per attempt. Big number takes more time but increases the
	// throughput.
	DefaultTLSCertCachePrune = 500
)

// ServerOpts is the datastructure to configure Server instance. The
// list of configuration options can be huge so it worth to use a
// separate object for this purpose.
type ServerOpts struct {
	// Concurrency is the number of concurrent connections to the service.
	Concurrency int

	// ReadBufferSize is the size of the read buffer for server TCP
	// connection.
	ReadBufferSize int

	// WriteBufferSize is the size of the write buffer for server TCP
	// connection.
	WriteBufferSize int

	// ReadTimeout is the timeout server uses to read from user request.
	ReadTimeout time.Duration

	// ReadTimeout is the timeout server uses to write to the user request.
	WriteTimeout time.Duration

	// CertCA is CA certificate for generating TLS certificate.
	CertCA []byte

	// CertKey is CA private key for generating TLS cerificate.
	CertKey []byte

	// OrganizationName is the name of organization which generated TLS
	// certificate.
	OrganizationName string

	// TLSCertCacheSize is the number of items we store in LRU cache
	// of generated TLS certificates before starting to prune obsoletes.
	TLSCertCacheSize int

	// TLSCertCachePrune is the number of items we prune from LRU
	// cache per attempt. Big number takes more time but increases the
	// throughput.
	TLSCertCachePrune int
}

// GetConcurrency returns a number of concurrent connections for the
// service we have to set.
func (s *ServerOpts) GetConcurrency() int {
	if s.Concurrency == 0 {
		return DefaultConcurrency
	}
	return s.Concurrency
}

// GetReadBufferSize returns the size of read buffer for the client
// connection in bytes we have to set to the server.
func (s *ServerOpts) GetReadBufferSize() int {
	if s.ReadBufferSize == 0 {
		return DefaultReadBufferSize
	}
	return s.ReadBufferSize
}

// GetWriteBufferSize returns the size of write buffer for the client
// connection in bytes we have to set to the server.
func (s *ServerOpts) GetWriteBufferSize() int {
	if s.WriteBufferSize == 0 {
		return DefaultWriteBufferSize
	}
	return s.WriteBufferSize
}

// GetReadTimeout is the timeout server has to use to read from user
// request.
func (s *ServerOpts) GetReadTimeout() time.Duration {
	if s.ReadTimeout == 0 {
		return DefaultReadTimeout
	}
	return s.ReadTimeout
}

// GetWriteTimeout is the timeout server has to use to write to user.
func (s *ServerOpts) GetWriteTimeout() time.Duration {
	if s.WriteTimeout == 0 {
		return DefaultWriteTimeout
	}
	return s.WriteTimeout
}

// GetCertCA returns a CA certificate which should be used to generate
// TLS certificates.
func (s *ServerOpts) GetCertCA() []byte {
	return s.CertCA
}

// GetCertKey returns a CA private key which should be used to generate
// TLS certificates.
func (s *ServerOpts) GetCertKey() []byte {
	return s.CertKey
}

// GetOrganizationName returns a name of the organization which should
// issue TLS certificates.
func (s *ServerOpts) GetOrganizationName() string {
	return s.OrganizationName
}

// GetTLSCertCacheSize returns the number of items to store in LRU cache
// of generated TLS certificates before starting to prune obsoletes.
func (s *ServerOpts) GetTLSCertCacheSize() int64 {
	return int64(s.TLSCertCacheSize)
}

// GetTLSCertCachePrune returns the number of items to prune from LRU
// cache per attempt.
func (s *ServerOpts) GetTLSCertCachePrune() uint32 {
	return uint32(s.TLSCertCachePrune)
}
