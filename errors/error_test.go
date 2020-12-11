package errors_test

import (
	"io"
	"testing"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/stretchr/testify/suite"
)

type ErrorTestSuite struct {
	suite.Suite

    e *errors.Error
}

func (suite *ErrorTestSuite) SetupTest() {
    suite.e = nil
}

func (suite *ErrorTestSuite) TestErrorMessage() {
	suite.Equal("", suite.e.Error())

	suite.e = &errors.Error{
		Err: errors.New("EOF"),
	}

	suite.Equal("EOF", suite.e.Error())

	suite.e.Message = "MyMessage"

	suite.Contains(suite.e.Error(), "MyMessage")
	suite.Contains(suite.e.Error(), "EOF")

	suite.e = &errors.Error{
		Message: "Another",
		Err:     suite.e,
	}

	suite.Contains(suite.e.Error(), "MyMessage")
	suite.Contains(suite.e.Error(), "EOF")
	suite.Contains(suite.e.Error(), "Another")

	suite.e = &errors.Error{
		Message: "QQ",
	}

	suite.Equal("QQ", suite.e.Error())
}

func (suite *ErrorTestSuite) TestUnwrap() {
	suite.Nil(suite.e.Unwrap())

	suite.e = &errors.Error{
		Err: io.EOF,
	}

	suite.True(errors.Is(suite.e.Unwrap(), io.EOF))
}

func (suite *ErrorTestSuite) TestGetStatusCode() {
    suite.Equal(0, suite.e.GetStatusCode())

    suite.e = &errors.Error{
        StatusCode: 1,
    }

    suite.Equal(1, suite.e.GetStatusCode())
}

func TestError(t *testing.T) {
	suite.Run(t, &ErrorTestSuite{})
}
