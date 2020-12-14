package layers_test

import (
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/suite"
)

type FilterSubnetLayerTestSuite struct {
	BaseLayerTestSuite
}

func (suite *FilterSubnetLayerTestSuite) SetupTest() {
	suite.BaseLayerTestSuite.SetupTest()

}

func (suite *FilterSubnetLayerTestSuite) TestIPv4() {
	suite.l, _ = layers.NewFilterSubnetsLayer([]net.IPNet{
		net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(24, 32)},
	})

	suite.Error(suite.l.OnRequest(suite.ctx))

	suite.l, _ = layers.NewFilterSubnetsLayer([]net.IPNet{
		net.IPNet{IP: net.ParseIP("10.0.0.10"), Mask: net.CIDRMask(24, 32)},
	})

	suite.NoError(suite.l.OnRequest(suite.ctx))
}

func (suite *FilterSubnetLayerTestSuite) TestIPv6() {
	_, netv6, _ := net.ParseCIDR("2001:db8:85a3:8d3:1319:8a2e:370:7348/64")
	suite.l, _ = layers.NewFilterSubnetsLayer([]net.IPNet{*netv6})

	suite.NoError(suite.l.OnRequest(suite.ctx))
}

func TestFilterSubnetLayer(t *testing.T) {
	suite.Run(t, &FilterSubnetLayerTestSuite{})
}
