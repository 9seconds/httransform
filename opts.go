package httransform

import (
	"time"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"
)

const (
	// DefaultConcurrency defines a default number of concurrently
	// managed requests processed by a proxy.
	DefaultConcurrency = 65536

	// DefaultReadBufferSize defines a default size of the buffer to use
	// on reading client requests.
	DefaultReadBufferSize = 16 * 1024

	// DefaultWriteBufferSize defines a default size of the buffer to
	// use on writing to client socket.
	DefaultWriteBufferSize = 16 * 1024

	// DefaultReadTimeout defines a default timeout for reading of
	// client request.
	DefaultReadTimeout = time.Minute

	// DefaultWriteTimeout defines a default timeout for writing
	// to client.
	DefaultWriteTimeout = time.Minute

	// DefaultTCPKeepAlivePeriod defines a default time period for TCP
	// keepalive probes.
	DefaultTCPKeepAlivePeriod = 30 * time.Second

	// DefaultMaxRequestBodySize defines a max size of the request body.
	DefaultMaxRequestBodySize = 1024 * 1024 * 100
)

// ServerOpts defines server options to use.
//
// Please pay attention that each field is optional. We do provide sane
// defaults if you do not want to pass anything there.
type ServerOpts struct {
	// Concurrency defines a number of concurrently managed requests.
	Concurrency uint

	// ReadBufferSize defines a size of the buffer allocated for reading
	// from client socket.
	ReadBufferSize uint

	// WriteBufferSize defines a size of the buffer allocated for
	// writing to a client socket.
	WriteBufferSize uint

	// MaxRequestBodySize defines a max size of the request body.
	MaxRequestBodySize uint

	// ReadTimeout defines a timeout for reading from client socket.
	ReadTimeout time.Duration

	// WriteTimeout defines a timeout for writing to client socket.
	WriteTimeout time.Duration

	// TCPKeepAlivePeriod defines a time period between 2 consecutive
	// TCP keepalive probes.
	TCPKeepAlivePeriod time.Duration

	// EventProcessorFactory defines a factory method which produces
	// event processors.
	EventProcessorFactory events.ProcessorFactory

	// TLSCertCA is a bytes which contains TLS CA certificate. This
	// certificate is required for generating fake TLS certifiates for
	// websites on TLS connection upgrades.
	TLSCertCA []byte

	// TLSPrivateKey is a bytes which contains TLS private key. This
	// certificate is required for generating fake TLS certifiates for
	// websites on TLS connection upgrades.
	TLSPrivateKey []byte

	// Layers defines a list of layers, middleware which should be used
	// by proxy.
	Layers []layers.Layer

	// Executor defines an executor function which should be used
	// to terminate HTTP request and fill HTTP response.
	Executor executor.Executor

	// Authenticator is an interface which is used to authenticate a
	// request.
	Authenticator auth.Interface

	// TLSSkipVerify defines if we need to verify TLS certifiates we have
	// to deal with.
	TLSSkipVerify bool
}

// GetConcurrency returns a concurrency paying attention to default
// value.
func (s *ServerOpts) GetConcurrency() int {
	if s == nil || s.Concurrency == 0 {
		return DefaultConcurrency
	}

	return int(s.Concurrency)
}

// GetReadBufferSize returns a read buffer size paying attention to
// default value.
func (s *ServerOpts) GetReadBufferSize() int {
	if s == nil || s.ReadBufferSize == 0 {
		return DefaultReadBufferSize
	}

	return int(s.ReadBufferSize)
}

// GetWriteBufferSize returns a write buffer size paying attention to
// default value.
func (s *ServerOpts) GetWriteBufferSize() int {
	if s == nil || s.WriteBufferSize == 0 {
		return DefaultWriteBufferSize
	}

	return int(s.WriteBufferSize)
}

// GetReadTimeout returns a read timeout paying attention to default
// value.
func (s *ServerOpts) GetReadTimeout() time.Duration {
	if s == nil || s.ReadTimeout == 0 {
		return DefaultReadTimeout
	}

	return s.ReadTimeout
}

// GetWriteTimeout returns a write timeout paying attention to default
// value.
func (s *ServerOpts) GetWriteTimeout() time.Duration {
	if s == nil || s.WriteTimeout == 0 {
		return DefaultWriteTimeout
	}

	return s.WriteTimeout
}

// GetTCPKeepAlivePeriod returns a period for TCP keepalive probes
// paying attention to default value.
func (s *ServerOpts) GetTCPKeepAlivePeriod() time.Duration {
	if s == nil || s.TCPKeepAlivePeriod == 0 {
		return DefaultTCPKeepAlivePeriod
	}

	return s.TCPKeepAlivePeriod
}

// GetMaxRequestBodySize returns max request body size paying attention
// to default value.
func (s *ServerOpts) GetMaxRequestBodySize() int {
	if s == nil || s.MaxRequestBodySize == 0 {
		return DefaultMaxRequestBodySize
	}

	return int(s.MaxRequestBodySize)
}

// GetEventProcessorFactory returns an event factory paying attention to
// default value.
func (s *ServerOpts) GetEventProcessorFactory() events.ProcessorFactory {
	if s == nil || s.EventProcessorFactory == nil {
		return events.NoopProcessorFactory
	}

	return s.EventProcessorFactory
}

// GetTLSCertCA returns a given TLS CA certificate.
func (s *ServerOpts) GetTLSCertCA() []byte {
	if s == nil {
		return nil
	}

	return s.TLSCertCA
}

// GetTLSPrivateKey returns a given TLS private key.
func (s *ServerOpts) GetTLSPrivateKey() []byte {
	if s == nil {
		return nil
	}

	return s.TLSPrivateKey
}

// GetTLSSkipVerify returns a sign if we need to skip TLS verification.
func (s *ServerOpts) GetTLSSkipVerify() bool {
	return s != nil && s.TLSSkipVerify
}

// GetLayers returns a set of server layers to use.
func (s *ServerOpts) GetLayers() []layers.Layer {
	toReturn := []layers.Layer{layerStartHeaders{}}

	if s != nil {
		toReturn = append(toReturn, s.Layers...)
	}

	toReturn = append(toReturn, layerFinishHeaders{})

	return toReturn
}

// GetLayers returns an authenticator instance to use paying attention
// to default value (no auth).
func (s *ServerOpts) GetAuthenticator() auth.Interface {
	if s == nil || s.Authenticator == nil {
		return auth.NoopAuth{}
	}

	return s.Authenticator
}

// GetExecutor returns an instance of executor to use.
func (s *ServerOpts) GetExecutor() executor.Executor {
	if s == nil {
		return nil
	}

	return s.Executor
}
