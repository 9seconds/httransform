package ca

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/stretchr/testify/suite"
)

type TLSConfigTestSuite struct {
	suite.Suite

	conf  *tls.Config
	cache *ccache.Cache
}

func (suite *TLSConfigTestSuite) SetupTest() {
	suite.conf = &tls.Config{}
	suite.cache = ccache.New(ccache.Configure())
	suite.cache.Set("key", suite.conf, time.Minute)
}

func (suite *TLSConfigTestSuite) TearDownTest() {
	suite.cache.Stop()
}

func (suite *TLSConfigTestSuite) TestGetReleaseGet() {
	item := suite.cache.TrackingGet("key")
	suite.NotEqual(item, ccache.NilTracked)

	conf := TLSConfig{item}
	suite.Equal(conf.Get(), suite.conf)

	conf.Release()
	conf.Release()
	suite.Nil(conf.Get())
}

func TestTLSConfig(t *testing.T) {
	suite.Run(t, &TLSConfigTestSuite{})
}
