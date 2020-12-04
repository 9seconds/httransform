package dialers_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/stretchr/testify/suite"
)

type DialerFromURLTestSuite struct {
	suite.Suite
}

func (suite *DialerFromURLTestSuite) TestIncorrectURL() {
	_, err := dialers.DialerFromURL(dialers.Opts{}, "")

	suite.Error(err)
}

func (suite *DialerFromURLTestSuite) TestSOCKS5() {
	_, err := dialers.DialerFromURL(dialers.Opts{}, "socks5://addr.com:3131")

	suite.NoError(err)
}

func (suite *DialerFromURLTestSuite) TestHTTP() {
	_, err := dialers.DialerFromURL(dialers.Opts{}, "http://user:password@addr.com:3131")

	suite.NoError(err)
}

func (suite *DialerFromURLTestSuite) TestHTTPS() {
	_, err := dialers.DialerFromURL(dialers.Opts{}, "https://user:password@addr.com:3131")

	suite.NoError(err)
}

func (suite *DialerFromURLTestSuite) TestUnkown() {
	_, err := dialers.DialerFromURL(dialers.Opts{}, "unknown://user:password@addr.com:3131")

	suite.Error(err)
}

func TestDialerFromURL(t *testing.T) {
	suite.Run(t, &DialerFromURLTestSuite{})
}
