package httransform

import (
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type BaseLayerStateTestSuite struct {
	suite.Suite

	state *LayerState
}

func (suite *BaseLayerStateTestSuite) SetupTest() {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	ctx := &fasthttp.RequestCtx{
		Request:  *req,
		Response: *resp,
	}
	requestHeaderSet := getHeaderSet()
	responseHeaderSet := getHeaderSet()
	user := []byte("user")
	password := []byte("password")

	suite.state = getLayerState()
	initLayerState(suite.state, ctx, requestHeaderSet, responseHeaderSet, true, user, password)
}

type LayerStateTestSuite struct {
	BaseLayerStateTestSuite
}

func (suite *LayerStateTestSuite) TestGetSet() {
	_, ok := suite.state.Get("key")
	suite.False(ok)

	suite.state.Set("key", "value")
	val, ok := suite.state.Get("key")
	suite.True(ok)
	suite.Equal(val.(string), "value")
}

type ConnectionCloseLayerTestSuite struct {
	BaseLayerStateTestSuite

	layer *ConnectionCloseLayer
}

func (suite *ConnectionCloseLayerTestSuite) SetupTest() {
	suite.BaseLayerStateTestSuite.SetupTest()

	suite.layer = &ConnectionCloseLayer{}
}

func (suite *ConnectionCloseLayerTestSuite) TestOnRequestHasHeader() {
	suite.state.RequestHeaders.SetString("Connection", "close")

	suite.Nil(suite.layer.OnRequest(suite.state))
	value, ok := suite.state.RequestHeaders.GetString("connection")
	suite.True(ok)
	suite.Equal(value, "close")
}

func (suite *ConnectionCloseLayerTestSuite) TestOnRequestNoHeader() {
	suite.Nil(suite.layer.OnRequest(suite.state))
	_, ok := suite.state.RequestHeaders.GetString("connection")
	suite.False(ok)
}

func (suite *ConnectionCloseLayerTestSuite) TestOnResponseHasHeaderNoError() {
	suite.state.ResponseHeaders.SetString("Connection", "keep-alive")

	suite.layer.OnResponse(suite.state, nil)
	value, ok := suite.state.ResponseHeaders.GetString("connection")
	suite.True(ok)
	suite.Equal(value, "close")

	value, ok = suite.state.ResponseHeaders.GetString("proxy-connection")
	suite.True(ok)
	suite.Equal(value, "close")
}

func (suite *ConnectionCloseLayerTestSuite) TestOnResponseHasHeaderError() {
	suite.state.ResponseHeaders.SetString("Connection", "keep-alive")

	suite.layer.OnResponse(suite.state, errors.New("Some error"))
	value, ok := suite.state.ResponseHeaders.GetString("connection")
	suite.True(ok)
	suite.Equal(value, "close")

	value, ok = suite.state.ResponseHeaders.GetString("proxy-connection")
	suite.True(ok)
	suite.Equal(value, "close")
}

func (suite *ConnectionCloseLayerTestSuite) TestOnResponseNoHeaderError() {
	suite.layer.OnResponse(suite.state, errors.New("Some error"))

	value, ok := suite.state.ResponseHeaders.GetString("connection")
	suite.True(ok)
	suite.Equal(value, "close")

	value, ok = suite.state.ResponseHeaders.GetString("proxy-connection")
	suite.True(ok)
	suite.Equal(value, "close")
}

type ProxyHeadersLayerTestSuite struct {
	BaseLayerStateTestSuite

	layer           *ProxyHeadersLayer
	hopByHopHeaders []string
}

func (suite *ProxyHeadersLayerTestSuite) SetupTest() {
	suite.BaseLayerStateTestSuite.SetupTest()

	suite.layer = &ProxyHeadersLayer{}
	suite.hopByHopHeaders = []string{
		"Proxy-Connection",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Connection",
		"Keep-Alive",
		"TE",
		"Trailers",
	}
}

func (suite *ProxyHeadersLayerTestSuite) TestOnRequest() {
	for _, header := range suite.hopByHopHeaders {
		suite.state.RequestHeaders.SetString(header, "value")
	}
	suite.Nil(suite.layer.OnRequest(suite.state))

	for _, header := range suite.hopByHopHeaders {
		_, ok := suite.state.RequestHeaders.GetString(header)
		suite.False(ok, header)
	}
}

func (suite *ProxyHeadersLayerTestSuite) TestOnResponseError() {
	for _, header := range suite.hopByHopHeaders {
		suite.state.ResponseHeaders.SetString(header, "value")
	}
	suite.layer.OnResponse(suite.state, errors.New("error"))

	for _, header := range suite.hopByHopHeaders {
		_, ok := suite.state.ResponseHeaders.GetString(header)
		suite.False(ok, header)
	}
}

func (suite *ProxyHeadersLayerTestSuite) TestOnResponseNoError() {
	for _, header := range suite.hopByHopHeaders {
		suite.state.ResponseHeaders.SetString(header, "value")
	}
	suite.layer.OnResponse(suite.state, nil)

	for _, header := range suite.hopByHopHeaders {
		_, ok := suite.state.ResponseHeaders.GetString(header)
		suite.False(ok, header)
	}
}

func TestLayerState(t *testing.T) {
	suite.Run(t, &LayerStateTestSuite{})
}

func TestConnectionCloseLayer(t *testing.T) {
	suite.Run(t, &ConnectionCloseLayerTestSuite{})
}

func TestProxyHeadersLayer(t *testing.T) {
	suite.Run(t, &ProxyHeadersLayerTestSuite{})
}
