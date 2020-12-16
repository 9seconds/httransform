package httransform_test

import (
	"testing"
	"time"

	"github.com/9seconds/httransform/v2"
	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/suite"
)

type OptsTestSuite struct {
	suite.Suite

	o httransform.ServerOpts
}

func (suite *OptsTestSuite) SetupTest() {
	suite.o = httransform.ServerOpts{}
}

func (suite *OptsTestSuite) TestNil() {
	var opts *httransform.ServerOpts

	suite.Equal(httransform.DefaultConcurrency, opts.GetConcurrency())
	suite.Equal(httransform.DefaultReadBufferSize, opts.GetReadBufferSize())
	suite.Equal(httransform.DefaultWriteBufferSize, opts.GetWriteBufferSize())
	suite.Equal(httransform.DefaultReadTimeout, opts.GetReadTimeout())
	suite.Equal(httransform.DefaultWriteTimeout, opts.GetWriteTimeout())
	suite.Equal(httransform.DefaultTCPKeepAlivePeriod, opts.GetTCPKeepAlivePeriod())
	suite.NotNil(opts.GetEventProcessorFactory())
	suite.Empty(opts.GetTLSCertCA())
	suite.Empty(opts.GetTLSPrivateKey())
	suite.False(opts.GetTLSSkipVerify())
	suite.Len(opts.GetLayers(), 2)
	suite.IsType(auth.NoopAuth{}, opts.GetAuthenticator())
	suite.Nil(opts.GetExecutor())
}

func (suite *OptsTestSuite) TestGetConcurrency() {
	suite.Equal(httransform.DefaultConcurrency, suite.o.GetConcurrency())

	suite.o.Concurrency = httransform.DefaultConcurrency + 1

	suite.Equal(httransform.DefaultConcurrency+1, suite.o.GetConcurrency())
}

func (suite *OptsTestSuite) TestGetReadBufferSize() {
	suite.Equal(httransform.DefaultReadBufferSize, suite.o.GetReadBufferSize())

	suite.o.ReadBufferSize = httransform.DefaultReadBufferSize + 1

	suite.Equal(httransform.DefaultReadBufferSize+1, suite.o.GetReadBufferSize())
}

func (suite *OptsTestSuite) TestGetWriteBufferSize() {
	suite.Equal(httransform.DefaultWriteBufferSize, suite.o.GetWriteBufferSize())

	suite.o.WriteBufferSize = httransform.DefaultWriteBufferSize + 1

	suite.Equal(httransform.DefaultWriteBufferSize+1, suite.o.GetWriteBufferSize())
}

func (suite *OptsTestSuite) TestGetReadTimeout() {
	suite.Equal(httransform.DefaultReadTimeout, suite.o.GetReadTimeout())

	suite.o.ReadTimeout = httransform.DefaultReadTimeout + time.Second

	suite.Equal(httransform.DefaultReadTimeout+time.Second, suite.o.GetReadTimeout())
}

func (suite *OptsTestSuite) TestGetWriteTimeout() {
	suite.Equal(httransform.DefaultWriteTimeout, suite.o.GetWriteTimeout())

	suite.o.WriteTimeout = httransform.DefaultWriteTimeout + time.Second

	suite.Equal(httransform.DefaultWriteTimeout+time.Second, suite.o.GetWriteTimeout())
}

func (suite *OptsTestSuite) TestGetTCPKeepAlivePeriod() {
	suite.Equal(httransform.DefaultTCPKeepAlivePeriod, suite.o.GetTCPKeepAlivePeriod())

	suite.o.TCPKeepAlivePeriod = httransform.DefaultTCPKeepAlivePeriod + time.Second

	suite.Equal(httransform.DefaultTCPKeepAlivePeriod+time.Second, suite.o.GetTCPKeepAlivePeriod())
}

func (suite *OptsTestSuite) TestGetMaxRequestBodySize() {
	suite.Equal(httransform.DefaultMaxRequestBodySize, suite.o.GetMaxRequestBodySize())

	suite.o.MaxRequestBodySize = httransform.DefaultMaxRequestBodySize + 1

	suite.Equal(httransform.DefaultMaxRequestBodySize+1, suite.o.GetMaxRequestBodySize())
}

func (suite *OptsTestSuite) TestGetEventProcessorFactory() {
	suite.NotNil(suite.o.GetEventProcessorFactory())

	suite.o.EventProcessorFactory = func() events.Processor { return nil }

	// https://github.com/stretchr/testify/issues/182
	suite.NotNil(suite.o.GetEventProcessorFactory())
}

func (suite *OptsTestSuite) TestGetTLSCertCA() {
	suite.Empty(suite.o.GetTLSCertCA())

	suite.o.TLSCertCA = []byte("hello")

	suite.Equal("hello", string(suite.o.GetTLSCertCA()))
}

func (suite *OptsTestSuite) TestGetTLSPrivateKey() {
	suite.Empty(suite.o.GetTLSPrivateKey())

	suite.o.TLSPrivateKey = []byte("hello")

	suite.Equal("hello", string(suite.o.GetTLSPrivateKey()))
}

func (suite *OptsTestSuite) TestGetTLSSkipVerify() {
	suite.False(suite.o.GetTLSSkipVerify())

	suite.o.TLSSkipVerify = true

	suite.True(suite.o.GetTLSSkipVerify())
}

func (suite *OptsTestSuite) TestGetLayers() {
	suite.Len(suite.o.GetLayers(), 2)

    lr := layers.TimeoutLayer{Timeout: time.Second}

	suite.o.Layers = []layers.Layer{lr}

	suite.Len(suite.o.GetLayers(), 3)
    suite.Equal(lr, suite.o.GetLayers()[1])
}

func (suite *OptsTestSuite) TestGetAuthenticator() {
    suite.IsType(auth.NoopAuth{}, suite.o.GetAuthenticator())

    suite.o.Authenticator = auth.NewBasicAuth(nil)

    suite.Equal(suite.o.Authenticator, suite.o.GetAuthenticator())
}

func (suite *OptsTestSuite) TestGetExecutor() {
    suite.Nil(suite.o.GetExecutor())

    suite.o.Executor = func(_ *layers.Context) error { return nil }

    suite.NotNil(suite.o.GetExecutor())
}

func TestOpts(t *testing.T) {
	suite.Run(t, &OptsTestSuite{})
}
