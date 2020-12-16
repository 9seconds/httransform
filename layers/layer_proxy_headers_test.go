package layers_test

import (
	"io"
	"testing"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/suite"
)

type LayerProxyHeadersTestSuite struct {
	BaseLayerTestSuite
}

func (suite *LayerProxyHeadersTestSuite) SetupTest() {
	suite.BaseLayerTestSuite.SetupTest()

	suite.l = layers.ProxyHeadersLayer{}
}

func (suite *LayerProxyHeadersTestSuite) TestOk() {
	suite.ctx.RequestHeaders.Append("proxy-authenticate", "111")
	suite.ctx.RequestHeaders.Append("proxy-Authenticate", "222")
	suite.ctx.RequestHeaders.Append("PROXY-AUTHORIZATION", "333")
	suite.ctx.ResponseHeaders.Append("Proxy-ConnECTION", "close")

	suite.NoError(suite.l.OnRequest(suite.ctx))
	suite.NoError(suite.l.OnResponse(suite.ctx, nil))

	suite.Empty(suite.ctx.RequestHeaders.GetAll("Proxy-Authenticate"))
	suite.Empty(suite.ctx.RequestHeaders.GetAll("Proxy-Authorization"))
	suite.Empty(suite.ctx.ResponseHeaders.GetAll("Proxy-Connection"))
}

func (suite *LayerProxyHeadersTestSuite) TestErr() {
	suite.ctx.RequestHeaders.Append("proxy-authenticate", "111")
	suite.ctx.RequestHeaders.Append("proxy-Authenticate", "222")
	suite.ctx.RequestHeaders.Append("PROXY-AUTHORIZATION", "333")
	suite.ctx.ResponseHeaders.Append("Proxy-ConnECTION", "close")

	suite.NoError(suite.l.OnRequest(suite.ctx))
	suite.Equal(io.EOF, suite.l.OnResponse(suite.ctx, io.EOF))

	suite.Empty(suite.ctx.RequestHeaders.GetAll("Proxy-Authenticate"))
	suite.Empty(suite.ctx.RequestHeaders.GetAll("Proxy-Authorization"))
	suite.Empty(suite.ctx.ResponseHeaders.GetAll("Proxy-Connection"))
}

func TestLayerProxyHeaders(t *testing.T) {
	suite.Run(t, &LayerProxyHeadersTestSuite{})
}
