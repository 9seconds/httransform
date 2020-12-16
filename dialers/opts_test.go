package dialers_test

import (
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/stretchr/testify/suite"
)

type OptsTestSuite struct {
	suite.Suite
}

func (suite *OptsTestSuite) TestGetTimeout() {
	opt := dialers.Opts{}

	suite.Equal(dialers.DefaultTimeout, opt.GetTimeout())

	opt.Timeout = time.Minute

	suite.Equal(time.Minute, opt.GetTimeout())
}

func (suite *OptsTestSuite) TestGetTLSSkipVerify() {
	opt := dialers.Opts{}

	suite.False(opt.GetTLSSkipVerify())

	opt.TLSSkipVerify = true

	suite.True(opt.GetTLSSkipVerify())
}

func TestOpts(t *testing.T) {
	suite.Run(t, &OptsTestSuite{})
}
