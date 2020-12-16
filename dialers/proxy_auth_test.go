package dialers_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/stretchr/testify/suite"
)

type ProxyAuthTestSuite struct {
	suite.Suite
}

func (suite *ProxyAuthTestSuite) TestIncorrectAddress() {
	_, err := dialers.NewProxyAuth("address.com", "", "")

	suite.Error(err)
}

func (suite *ProxyAuthTestSuite) TestLocalhost() {
	auth, err := dialers.NewProxyAuth(":80", "", "")

	suite.NoError(err)
	suite.Equal("127.0.0.1:80", auth.Address)
	suite.NotEmpty(auth.String())
}

func (suite *ProxyAuthTestSuite) TestNoCredentials() {
	auth, err := dialers.NewProxyAuth("address.com:3128", "", "")

	suite.NoError(err)
	suite.False(auth.HasCredentials())
}

func (suite *ProxyAuthTestSuite) TestHostPort() {
	auth, err := dialers.NewProxyAuth("address.com:3128", "", "")

	suite.NoError(err)
	suite.Equal("address.com", auth.Host())
	suite.Equal(3128, auth.Port())
}

func (suite *ProxyAuthTestSuite) TestCredentials() {
	auth, err := dialers.NewProxyAuth("address.com:3128", "user", "password")

	suite.NoError(err)
	suite.Equal("user", auth.Username)
	suite.Equal("password", auth.Password)
	suite.True(auth.HasCredentials())
}

func (suite *ProxyAuthTestSuite) TestCredentialsOnlySingle() {
	auth, err := dialers.NewProxyAuth("address.com:3128", "user", "")

	suite.NoError(err)
	suite.True(auth.HasCredentials())

	auth, err = dialers.NewProxyAuth("address.com:3128", "", "password")

	suite.NoError(err)
	suite.True(auth.HasCredentials())
}

func TestProxyAuth(t *testing.T) {
	suite.Run(t, &ProxyAuthTestSuite{})
}
