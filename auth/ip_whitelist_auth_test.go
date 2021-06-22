package auth_test

import (
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type IPWhitelistAuthTestSuite struct {
	suite.Suite

	req  *fasthttp.Request
	ctx  *fasthttp.RequestCtx
	auth auth.Interface
}

func (suite *IPWhitelistAuthTestSuite) SetupTest() {
	suite.req = &fasthttp.Request{}

	suite.req.SetRequestURI("http://example.com/image.gif")
	suite.req.Header.SetMethod("GET")

	suite.ctx = &fasthttp.RequestCtx{}

	_, net127, _ := net.ParseCIDR("127.0.0.1/24")
	_, net192, _ := net.ParseCIDR("172.16.0.1/24")
	_, netv6, _ := net.ParseCIDR("2001:db8:85a3:8d3:1319:8a2e:370:7348/64")

	suite.auth, _ = auth.NewIPWhitelist(map[string][]net.IPNet{
		"user1": {*net127, *net192},
		"user2": {*netv6},
	})
}

func (suite *IPWhitelistAuthTestSuite) TestUnknown() {
	suite.ctx.Init(suite.req,
		&net.TCPAddr{IP: net.ParseIP("10.0.0.10"), Port: 9000},
		nil)

	_, err := suite.auth.Authenticate(suite.ctx)

	suite.EqualError(err, auth.ErrFailedAuth.Error())
}

func (suite *IPWhitelistAuthTestSuite) TestUser1() {
	suite.ctx.Init(suite.req,
		&net.TCPAddr{IP: net.ParseIP("172.16.0.36"), Port: 9000},
		nil)

	user, err := suite.auth.Authenticate(suite.ctx)

	suite.NoError(err)
	suite.Equal("user1", user)
}

func (suite *IPWhitelistAuthTestSuite) TestUser2() {
	suite.ctx.Init(suite.req,
		&net.TCPAddr{IP: net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348"), Port: 9000},
		nil)

	user, err := suite.auth.Authenticate(suite.ctx)

	suite.NoError(err)
	suite.Equal("user2", user)
}

func TestIpWhitelist(t *testing.T) {
	suite.Run(t, &IPWhitelistAuthTestSuite{})
}
