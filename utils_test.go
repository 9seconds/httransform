package httransform

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type MakeSimpleResponseTestSuite struct {
	suite.Suite

	resp *fasthttp.Response
}

func (suite *MakeSimpleResponseTestSuite) SetupTest() {
	suite.resp = fasthttp.AcquireResponse()
	suite.resp.Reset()
}

func (suite *MakeSimpleResponseTestSuite) TearDownTest() {
	fasthttp.ReleaseResponse(suite.resp)
}

func (suite *MakeSimpleResponseTestSuite) TestOverrideValues() {
	suite.resp.SetStatusCode(fasthttp.StatusMovedPermanently)
	suite.resp.SetBodyString("HELLO")
	suite.resp.SetConnectionClose()

	MakeSimpleResponse(suite.resp, "overriden", fasthttp.StatusOK)

	suite.Equal(suite.resp.StatusCode(), fasthttp.StatusOK)
	suite.False(suite.resp.ConnectionClose())
	suite.Equal(suite.resp.Body(), []byte("overriden"))
	suite.Equal(suite.resp.Header.ContentType(), []byte("text/plain"))
}

type ExtractAuthenticationTestSuite struct {
	suite.Suite
}

func (suite *ExtractAuthenticationTestSuite) TestEmpty() {
	_, _, err := ExtractAuthentication([]byte{})

	suite.NotNil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestGarbage() {
	_, _, err := ExtractAuthentication([]byte("adfjsfsjfhaskfsjfsjh"))

	suite.NotNil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestDigest() {
	_, _, err := ExtractAuthentication([]byte("Digest dXNlcjpwYXNz"))

	suite.NotNil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestLowerCasedType() {
	_, _, err := ExtractAuthentication([]byte("basic dXNlcjpwYXNz"))

	suite.NotNil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestIncorrectPayload() {
	_, _, err := ExtractAuthentication([]byte("Basic XXXXXXXXX"))

	suite.NotNil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestNoPassword() {
	_, _, err := ExtractAuthentication([]byte("Basic dXNlcg=="))

	suite.NotNil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestEmptyPassword() {
	user, password, err := ExtractAuthentication([]byte("Basic dXNlcjo="))

	suite.Equal(user, []byte("user"))
	suite.Len(password, 0)
	suite.Nil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestEmptyUser() {
	user, password, err := ExtractAuthentication([]byte("Basic OnBhc3M="))

	suite.Equal(password, []byte("pass"))
	suite.Len(user, 0)
	suite.Nil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestAllEmpty() {
	user, password, err := ExtractAuthentication([]byte("Basic Og=="))

	suite.Len(user, 0)
	suite.Len(password, 0)
	suite.Nil(err)
}

func (suite *ExtractAuthenticationTestSuite) TestUserPass() {
	user, password, err := ExtractAuthentication([]byte("Basic dXNlcjpwYXNz"))

	suite.Equal(user, []byte("user"))
	suite.Equal(password, []byte("pass"))
	suite.Nil(err)
}

type MakeProxyAuthorizationHeaderValueTestSuite struct {
	suite.Suite
}

func (suite *MakeProxyAuthorizationHeaderValueTestSuite) TestEmpty() {
	result := MakeProxyAuthorizationHeaderValue(&url.URL{})

	suite.Len(result, 0)
}

func (suite *MakeProxyAuthorizationHeaderValueTestSuite) TestUserOnly() {
	result := MakeProxyAuthorizationHeaderValue(&url.URL{
		User: url.User("username"),
	})

	suite.Equal(result, []byte("Basic dXNlcm5hbWU6"))
}

func (suite *MakeProxyAuthorizationHeaderValueTestSuite) TestUserPass() {
	result := MakeProxyAuthorizationHeaderValue(&url.URL{
		User: url.UserPassword("username", "password"),
	})

	suite.Equal(result, []byte("Basic dXNlcm5hbWU6cGFzc3dvcmQ="))
}

func TestMakeSimpleResponse(t *testing.T) {
	suite.Run(t, &MakeSimpleResponseTestSuite{})
}

func TestExtractAuthentication(t *testing.T) {
	suite.Run(t, &ExtractAuthenticationTestSuite{})
}

func TestMakeProxuAuthorizationHeaderValue(t *testing.T) {
	suite.Run(t, &MakeProxyAuthorizationHeaderValueTestSuite{})
}
