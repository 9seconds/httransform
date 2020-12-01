package headers_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/stretchr/testify/suite"
)

type ValuesTestSuite struct {
	suite.Suite
}

func (suite *ValuesTestSuite) TestNothing() {
	suite.Empty(headers.Values(""))
}

func (suite *ValuesTestSuite) TestSingle() {
	suite.Equal([]string{"value"}, headers.Values("value"))
}

func (suite *ValuesTestSuite) TestMany() {
	suite.Equal([]string{"value", "hello"}, headers.Values("value, hello"))
}

func (suite *ValuesTestSuite) TestManySpace() {
	suite.Equal([]string{"value   ", "hello"}, headers.Values("value   , hello"))
}

func TestValues(t *testing.T) {
	suite.Run(t, &ValuesTestSuite{})
}
