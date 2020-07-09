package httransform

import (
	"time"

	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/layers"
)

const (
	DefaultConcurrency      = fasthttp.DefaultConcurrency
	DefaultReadBufferSize   = 4096
	DefaultWriteBufferSize  = 4096
	DefaultReadTimeout      = time.Minute
	DefaultWriteTimeout     = time.Minute
	DefaultTLSCertCacheSize = 10000
)

type ServerOpts struct {
	Concurrency      int
	ReadBufferSize   int
	WriteBufferSize  int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	CertCA           []byte
	CertKey          []byte
	OrganizationName string
	TLSCertCacheSize uint
	Layers           []layers.Layer
	Auth             auth.Auth
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

func (s *ServerOpts) GetTLSCertCacheSize() int {
	if s.TLSCertCacheSize == 0 {
		return DefaultTLSCertCacheSize
	}

	return int(s.TLSCertCacheSize)
}

func (s *ServerOpts) GetLayers() []layers.Layer {
	return s.Layers
}

func (s *ServerOpts) GetAuth() auth.Auth {
	if s.Auth == nil {
		return auth.NewNoopAuth()
	}

	return s.Auth
}
