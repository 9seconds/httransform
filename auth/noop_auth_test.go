package auth_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/auth"
	"github.com/stretchr/testify/suite"
)

type NoopAuthTestSuite struct {
	suite.Suite
}

func (suite *NoopAuthTestSuite) TestAuthenticate() {
	_, err := auth.NoopAuth{}.Authenticate(nil)

	suite.NoError(err)
}

func TestNoopAuth(t *testing.T) {
	suite.Run(t, &NoopAuthTestSuite{})
}
