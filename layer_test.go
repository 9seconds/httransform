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

type AddRemoveHeaderLayerTestSuite struct {
	BaseLayerStateTestSuite

	layer *AddRemoveHeaderLayer
}

func (suite *AddRemoveHeaderLayerTestSuite) SetupTest() {
	suite.BaseLayerStateTestSuite.SetupTest()
	suite.layer = &AddRemoveHeaderLayer{}
}

func (suite *AddRemoveHeaderLayerTestSuite) TestOnRequest() {
	suite.layer.FrozenRequestHeaders = map[string]string{
		"User-Agent": "curl/7.58.0",
	}
	suite.layer.DefaultRequestHeaders = map[string]string{
		"User-Agent":      "golang",
		"Accept-Encoding": "gzip, deflate",
		"X-Token":         "123456",
	}
	suite.layer.AbsentRequestHeaders = []string{"proxy-authorization"}

	suite.state.RequestHeaders.SetString("user-agent", "???")
	suite.state.RequestHeaders.SetString("accept-Encoding", "gzip")
	suite.state.RequestHeaders.SetString("Proxy-Authorization", "basic")
	suite.state.RequestHeaders.SetString("Authorization", "Basic OG==")

	suite.Nil(suite.layer.OnRequest(suite.state))
	suite.Len(suite.state.RequestHeaders.Items(), 4)

	value, ok := suite.state.RequestHeaders.GetString("user-agent")
	suite.True(ok)
	suite.Equal(value, "curl/7.58.0")

	value, ok = suite.state.RequestHeaders.GetString("accept-encoding")
	suite.True(ok)
	suite.Equal(value, "gzip")

	_, ok = suite.state.RequestHeaders.GetString("proxy-authorization")
	suite.False(ok)

	value, ok = suite.state.RequestHeaders.GetString("x-token")
	suite.True(ok)
	suite.Equal(value, "123456")
}

func (suite *AddRemoveHeaderLayerTestSuite) TestOnResponseNoError() {
	suite.layer.FrozenResponseHeaders = map[string]string{
		"User-Agent": "curl/7.58.0",
	}
	suite.layer.DefaultResponseHeaders = map[string]string{
		"User-Agent":      "golang",
		"Accept-Encoding": "gzip, deflate",
		"X-Token":         "123456",
	}
	suite.layer.AbsentResponseHeaders = []string{"proxy-authorization"}

	suite.state.ResponseHeaders.SetString("user-agent", "???")
	suite.state.ResponseHeaders.SetString("accept-Encoding", "gzip")
	suite.state.ResponseHeaders.SetString("Proxy-Authorization", "basic")
	suite.state.ResponseHeaders.SetString("Authorization", "Basic OG==")

	suite.layer.OnResponse(suite.state, nil)
	suite.Len(suite.state.ResponseHeaders.Items(), 4)

	value, ok := suite.state.ResponseHeaders.GetString("user-agent")
	suite.True(ok)
	suite.Equal(value, "curl/7.58.0")

	value, ok = suite.state.ResponseHeaders.GetString("accept-encoding")
	suite.True(ok)
	suite.Equal(value, "gzip")

	_, ok = suite.state.ResponseHeaders.GetString("proxy-authorization")
	suite.False(ok)

	value, ok = suite.state.ResponseHeaders.GetString("x-token")
	suite.True(ok)
	suite.Equal(value, "123456")
}

func (suite *AddRemoveHeaderLayerTestSuite) TestOnResponseError() {
	suite.layer.FrozenResponseHeaders = map[string]string{
		"User-Agent": "curl/7.58.0",
	}
	suite.layer.DefaultResponseHeaders = map[string]string{
		"User-Agent":      "golang",
		"Accept-Encoding": "gzip, deflate",
		"X-Token":         "123456",
	}
	suite.layer.AbsentResponseHeaders = []string{"proxy-authorization"}

	suite.state.ResponseHeaders.SetString("user-agent", "???")
	suite.state.ResponseHeaders.SetString("accept-Encoding", "gzip")
	suite.state.ResponseHeaders.SetString("Proxy-Authorization", "basic")
	suite.state.ResponseHeaders.SetString("Authorization", "Basic OG==")

	suite.layer.OnResponse(suite.state, errors.New("123"))
	suite.Len(suite.state.ResponseHeaders.Items(), 4)

	value, ok := suite.state.ResponseHeaders.GetString("user-agent")
	suite.True(ok)
	suite.Equal(value, "curl/7.58.0")

	value, ok = suite.state.ResponseHeaders.GetString("accept-encoding")
	suite.True(ok)
	suite.Equal(value, "gzip")

	_, ok = suite.state.ResponseHeaders.GetString("proxy-authorization")
	suite.False(ok)

	value, ok = suite.state.ResponseHeaders.GetString("x-token")
	suite.True(ok)
	suite.Equal(value, "123456")
}

type ProxyAuthorizationBasicLayerTestSuite struct {
	BaseLayerStateTestSuite

	layer *ProxyAuthorizationBasicLayer
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) SetupTest() {
	suite.BaseLayerStateTestSuite.SetupTest()

	suite.layer = &ProxyAuthorizationBasicLayer{}
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnRequestAllMismatch() {
	suite.Equal(suite.layer.OnRequest(suite.state), ErrProxyAuthorization)
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnRequestUserMismatch() {
	suite.layer.Password = suite.state.ProxyPassword
	suite.Equal(suite.layer.OnRequest(suite.state), ErrProxyAuthorization)
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnRequestPasswordMismatch() {
	suite.layer.User = suite.state.ProxyUser
	suite.Equal(suite.layer.OnRequest(suite.state), ErrProxyAuthorization)
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnRequestMatch() {
	suite.layer.User = suite.state.ProxyUser
	suite.layer.Password = suite.state.ProxyPassword
	suite.Nil(suite.layer.OnRequest(suite.state))
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnResponseAnotherError() {
	suite.layer.OnResponse(suite.state, errors.New("Another error"))

	suite.Equal(suite.state.Response.StatusCode(), fasthttp.StatusOK)
	_, ok := suite.state.ResponseHeaders.GetString("proxy-authenticate")
	suite.False(ok)
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnResponseErrorNoRealm() {
	suite.layer.OnResponse(suite.state, ErrProxyAuthorization)

	suite.Equal(suite.state.Response.StatusCode(), fasthttp.StatusProxyAuthRequired)
	value, ok := suite.state.ResponseHeaders.GetString("proxy-authenticate")
	suite.True(ok)
	suite.Equal(value, "Basic")
}

func (suite *ProxyAuthorizationBasicLayerTestSuite) TestOnResponseErrorRealm() {
	suite.layer.Realm = "My realm"
	suite.layer.OnResponse(suite.state, ErrProxyAuthorization)

	suite.Equal(suite.state.Response.StatusCode(), fasthttp.StatusProxyAuthRequired)
	value, ok := suite.state.ResponseHeaders.GetString("proxy-authenticate")
	suite.True(ok)
	suite.Equal(value, "Basic realm=\"My realm\"")
}

func TestLayerState(t *testing.T) {
	suite.Run(t, &LayerStateTestSuite{})
}

func TestAddRemoveHeaderLayer(t *testing.T) {
	suite.Run(t, &AddRemoveHeaderLayerTestSuite{})
}

func TestProxyAuthorizationBasicLayer(t *testing.T) {
	suite.Run(t, &ProxyAuthorizationBasicLayerTestSuite{})
}
