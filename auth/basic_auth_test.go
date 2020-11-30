package auth_test

import (
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type BasicAuthTestSuite struct {
	suite.Suite

	ctx  *fasthttp.RequestCtx
	auth auth.Interface
}

func (suite *BasicAuthTestSuite) SetupTest() {
	suite.ctx = &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}

	req.SetRequestURI("http://example.com/image.gif")
	req.Header.SetMethod("GET")

	suite.ctx.Init(req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}, nil)

	suite.auth = auth.NewBasicAuth(map[string]string{
		"user": "password",
	})
}

func (suite *BasicAuthTestSuite) TestEmpty() {
	_, err := suite.auth.Authenticate(suite.ctx)

	suite.EqualError(err, auth.ErrAuthRequired.Error())
}

func (suite *BasicAuthTestSuite) TestIncorrectHeader() {
	suite.ctx.Request.Header.Add("Proxy-authorization", "HAHAHA")

	_, err := suite.auth.Authenticate(suite.ctx)

	suite.EqualError(err, auth.ErrMalformedHeaderValue.Error())
}

func (suite *BasicAuthTestSuite) TestUnsupportedSchema() {
	suite.ctx.Request.Header.Add("Proxy-authorization", "Token Mtox")

	_, err := suite.auth.Authenticate(suite.ctx)

	suite.Error(err)
}

func (suite *BasicAuthTestSuite) TestIncorrectValue() {
	suite.ctx.Request.Header.Add("Proxy-authorization", "basic 111")

	_, err := suite.auth.Authenticate(suite.ctx)

	suite.Error(err)
}

func (suite *BasicAuthTestSuite) TestOk() {
	suite.ctx.Request.Header.Add("Proxy-authorization", "basic   dXNlcjpwYXNzd29yZA==")

	user, err := suite.auth.Authenticate(suite.ctx)

	suite.Equal("user", user)
	suite.NoError(err)
}

func TestBasicAuth(t *testing.T) {
	suite.Run(t, &BasicAuthTestSuite{})
}
