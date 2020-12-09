package errors_test

import (
	stderrors "errors"
	"io"
	"os"
	"testing"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
}

func (suite *APITestSuite) TestNew() {
	suite.Contains(errors.New("error1").Error(), "error1")
}

func (suite *APITestSuite) TestUnwrap() {
	err := stderrors.New("error1")
	wrapper := &errors.Error{
		Err: err,
	}

	suite.Equal(err, errors.Unwrap(wrapper))
	suite.Equal(err, stderrors.Unwrap(wrapper))
}

func (suite *APITestSuite) TestAs() {
	err := &os.PathError{
		Path: "/opt",
	}
	wrapper := &errors.Error{
		Err: err,
	}

	var actualError *os.PathError

	suite.True(errors.As(wrapper, &actualError))
	suite.Equal(err, actualError)
	suite.True(stderrors.As(wrapper, &actualError))
	suite.Equal(err, actualError)
}

func (suite *APITestSuite) TestIs() {
	wrapper := &errors.Error{
		Err: io.EOF,
	}

	suite.True(errors.Is(wrapper, io.EOF))
	suite.True(stderrors.Is(wrapper, io.EOF))
}

func (suite *APITestSuite) TestAnnotate() {
	wrapper := &errors.Error{
		Err: io.EOF,
	}
	another := errors.Annotate(wrapper, "another message", "code", 0)

	suite.True(errors.Is(another, wrapper))
	suite.True(stderrors.Is(another, wrapper))
}

func TestAPI(t *testing.T) {
	suite.Run(t, &APITestSuite{})
}
