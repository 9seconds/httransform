package httransform

import (
	"time"

	"github.com/valyala/fasthttp"
)

const (
	DefaultConcurrency       = fasthttp.DefaultConcurrency
	DefaultReadBufferSize    = 4096
	DefaultWriteBufferSize   = 4096
	DefaultReadTimeout       = time.Minute * 5
	DefaultWriteTimeout      = time.Minute * 5
	DefaultTLSCertCacheSize  = 10000
	DefaultTLSCertCachePrune = 500
)

type ServerOpts struct {
	Concurrency       int
	ReadBufferSize    int
	WriteBufferSize   int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	CertCA            []byte
	CertKey           []byte
	OrganizationName  string
	TLSCertCacheSize  int
	TLSCertCachePrune int
}

func (s *ServerOpts) GetConcurrency() int {
	if s.Concurrency == 0 {
		return DefaultConcurrency
	}
	return s.Concurrency
}

func (s *ServerOpts) GetReadBufferSize() int {
	if s.ReadBufferSize == 0 {
		return DefaultReadBufferSize
	}
	return s.ReadBufferSize
}

func (s *ServerOpts) GetWriteBufferSize() int {
	if s.WriteBufferSize == 0 {
		return DefaultWriteBufferSize
	}
	return s.WriteBufferSize
}

func (s *ServerOpts) GetReadTimeout() time.Duration {
	if s.ReadTimeout == 0 {
		return DefaultReadTimeout
	}
	return s.ReadTimeout
}

func (s *ServerOpts) GetWriteTimeout() time.Duration {
	if s.WriteTimeout == 0 {
		return DefaultWriteTimeout
	}
	return s.WriteTimeout
}

func (s *ServerOpts) GetCertCA() []byte {
	return s.CertCA
}

func (s *ServerOpts) GetCertKey() []byte {
	return s.CertKey
}

func (s *ServerOpts) GetOrganizationName() string {
	return s.OrganizationName
}

func (s *ServerOpts) GetTLSCertCacheSize() int64 {
	return int64(s.TLSCertCacheSize)
}

func (s *ServerOpts) GetTLSCertCachePrune() uint32 {
	return uint32(s.TLSCertCachePrune)
}
