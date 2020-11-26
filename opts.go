package main

import (
	"encoding/base64"
	"net/url"
	"time"

	"github.com/9seconds/httransform/v2/client"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"
)

const (
	DefaultConcurrency        = 65536
	DefaultReadBufferSize     = 16 * 1024
	DefaultWriteBufferSize    = 16 * 1024
	DefaultReadTimeout        = time.Minute
	DefaultWriteTimeout       = time.Minute
	DefaultTCPKeepAlivePeriod = 30 * time.Second
	DefaultMaxRequestBodySize = 1024 * 1024 * 100
	DefaultTLSCacheSize       = 512
)

type ServerOpts struct {
	Concurrency        uint
	ReadBufferSize     uint
	WriteBufferSize    uint
	MaxRequestBodySize uint
	TLSCacheSize       uint
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	TCPKeepAlivePeriod time.Duration
	EventProcessor     events.EventProcessor
	TLSCertCA          []byte
	TLSPrivateKey      []byte
	Layers             []layers.Layer
	Auth               *url.Userinfo
	Executor           executor.Executor
	Client             *client.Client
	TLSSkipVerify      bool
}

func (s *ServerOpts) GetConcurrency() int {
	if s.Concurrency == 0 {
		return DefaultConcurrency
	}

	return int(s.Concurrency)
}

func (s *ServerOpts) GetReadBufferSize() int {
	if s.ReadBufferSize == 0 {
		return DefaultReadBufferSize
	}

	return int(s.ReadBufferSize)
}

func (s *ServerOpts) GetWriteBufferSize() int {
	if s.WriteBufferSize == 0 {
		return DefaultWriteBufferSize
	}

	return int(s.WriteBufferSize)
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

func (s *ServerOpts) GetTCPKeepAlivePeriod() time.Duration {
	if s.TCPKeepAlivePeriod == 0 {
		return DefaultTCPKeepAlivePeriod
	}

	return s.TCPKeepAlivePeriod
}

func (s *ServerOpts) GetMaxRequestBodySize() int {
	if s.MaxRequestBodySize == 0 {
		return DefaultMaxRequestBodySize
	}

	return int(s.MaxRequestBodySize)
}

func (s *ServerOpts) GetTLSCacheSize() int {
	if s.TLSCacheSize == 0 {
		return DefaultTLSCacheSize
	}

	return int(s.TLSCacheSize)
}

func (s *ServerOpts) GetEventProcessor() events.EventProcessor {
	if s.EventProcessor == nil {
		return events.DefaultProcessor
	}

	return s.EventProcessor
}

func (s *ServerOpts) GetTLSCertCA() []byte {
	return s.TLSCertCA
}

func (s *ServerOpts) GetTLSPrivateKey() []byte {
	return s.TLSPrivateKey
}

func (s *ServerOpts) GetTLSSkipVerify() bool {
	return s.TLSSkipVerify
}

func (s *ServerOpts) GetLayers() []layers.Layer {
	toReturn := []layers.Layer{layerStartHeaders{}}
	toReturn = append(toReturn, s.Layers...)
	toReturn = append(toReturn, layerFinishHeaders{})

	return toReturn
}

func (s *ServerOpts) GetAuth() []byte {
	user := s.Auth.Username()
	password, _ := s.Auth.Password()

	if user == "" && password == "" {
		return nil
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(user + ":" + password))

	return []byte(encoded)
}

func (s *ServerOpts) GetExecutor() executor.Executor {
	return s.Executor
}
