package errors_test

import (
	"io"
	"testing"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/stretchr/testify/suite"
)

type ErrorTestSuite struct {
	suite.Suite
}

func (suite *ErrorTestSuite) TestErrorMessage() {
	var e *errors.Error

	suite.Equal("", e.Error())

	e = &errors.Error{
		Err: errors.New("EOF"),
	}

	suite.Equal("EOF", e.Error())

	e.Message = "MyMessage"

	suite.Contains(e.Error(), "MyMessage")
	suite.Contains(e.Error(), "EOF")

	e = &errors.Error{
		Message: "Another",
		Err:     e,
	}

	suite.Contains(e.Error(), "MyMessage")
	suite.Contains(e.Error(), "EOF")
	suite.Contains(e.Error(), "Another")

	e = &errors.Error{
		Message: "QQ",
	}

	suite.Equal("QQ", e.Error())
}

func (suite *ErrorTestSuite) TestUnwrap() {
	var e *errors.Error

	suite.Nil(e.Unwrap())

	e = &errors.Error{
		Err: io.EOF,
	}

	suite.True(errors.Is(e.Unwrap(), io.EOF))
}

func TestError(t *testing.T) {
	suite.Run(t, &ErrorTestSuite{})
}
