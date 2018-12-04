package httransform

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ServerOptsTestSuite struct {
	suite.Suite

	opts ServerOpts
}

func (suite *ServerOptsTestSuite) SetupTest() {
	suite.opts = ServerOpts{
		CertCA:           []byte("certca"),
		CertKey:          []byte("certkey"),
		OrganizationName: "name",
	}
}

func (suite *ServerOptsTestSuite) TestGetConcurrency() {
	suite.Equal(suite.opts.GetConcurrency(), DefaultConcurrency)
	suite.opts.Concurrency = DefaultConcurrency + 1
	suite.Equal(suite.opts.GetConcurrency(), suite.opts.Concurrency)
}

func (suite *ServerOptsTestSuite) TestGetReadBufferSize() {
	suite.Equal(suite.opts.GetReadBufferSize(), DefaultReadBufferSize)
	suite.opts.ReadBufferSize = DefaultReadBufferSize + 1
	suite.Equal(suite.opts.GetReadBufferSize(), suite.opts.ReadBufferSize)
}

func (suite *ServerOptsTestSuite) TestGetWriteBufferSize() {
	suite.Equal(suite.opts.GetWriteBufferSize(), DefaultWriteBufferSize)
	suite.opts.WriteBufferSize = DefaultWriteBufferSize + 1
	suite.Equal(suite.opts.GetWriteBufferSize(), suite.opts.WriteBufferSize)
}

func (suite *ServerOptsTestSuite) TestGetReadTimeout() {
	suite.Equal(suite.opts.GetReadTimeout(), DefaultReadTimeout)
	suite.opts.ReadTimeout = DefaultReadTimeout + time.Second
	suite.Equal(suite.opts.GetReadTimeout(), suite.opts.ReadTimeout)
}

func (suite *ServerOptsTestSuite) TestGetWriteTimeout() {
	suite.Equal(suite.opts.GetWriteTimeout(), DefaultWriteTimeout)
	suite.opts.WriteTimeout = DefaultWriteTimeout + time.Second
	suite.Equal(suite.opts.GetWriteTimeout(), suite.opts.WriteTimeout)
}

func (suite *ServerOptsTestSuite) TestGetCertCA() {
	suite.Equal(suite.opts.GetCertCA(), []byte("certca"))
}

func (suite *ServerOptsTestSuite) TestGetCertKey() {
	suite.Equal(suite.opts.GetCertKey(), []byte("certkey"))
}

func (suite *ServerOptsTestSuite) TestGetOrganizationName() {
	suite.Equal(suite.opts.GetOrganizationName(), "name")
}

func (suite *ServerOptsTestSuite) TestGetTLSCertCacheSize() {
	suite.Equal(suite.opts.GetTLSCertCacheSize(), int64(DefaultTLSCertCacheSize))
	suite.opts.TLSCertCacheSize = DefaultTLSCertCacheSize + 1
	suite.Equal(suite.opts.GetTLSCertCacheSize(), int64(suite.opts.TLSCertCacheSize))
}

func (suite *ServerOptsTestSuite) TestGetTLSCertCachePrune() {
	suite.Equal(suite.opts.GetTLSCertCachePrune(), uint32(DefaultTLSCertCachePrune))
	suite.opts.TLSCertCachePrune = DefaultTLSCertCachePrune + 1
	suite.Equal(suite.opts.GetTLSCertCachePrune(), uint32(suite.opts.TLSCertCachePrune))
}

func TestServerOpts(t *testing.T) {
	suite.Run(t, &ServerOptsTestSuite{})
}
