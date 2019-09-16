package httransform

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

type MockHTTPRequestExecutor struct {
	mock.Mock
}

func (m *MockHTTPRequestExecutor) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	args := m.Called(req, resp)
	resp.SetStatusCode(fasthttp.StatusNotFound)

	return args.Error(0)
}

func (m *MockHTTPRequestExecutor) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	args := m.Called(req, resp, timeout)
	resp.SetStatusCode(fasthttp.StatusNotFound)

	return args.Error(0)
}

type ExecuteRequestTestSuite struct {
	suite.Suite

	executor *MockHTTPRequestExecutor
	request  *fasthttp.Request
	response *fasthttp.Response
}

func (suite *ExecuteRequestTestSuite) SetupTest() {
	suite.executor = &MockHTTPRequestExecutor{}
	suite.request = fasthttp.AcquireRequest()
	suite.response = fasthttp.AcquireResponse()
}

func (suite *ExecuteRequestTestSuite) TearDownTest() {
	fasthttp.ReleaseRequest(suite.request)
	fasthttp.ReleaseResponse(suite.response)
}

func (suite *ExecuteRequestTestSuite) TestExecuteRequestOK() {
	suite.executor.On("Do", suite.request, suite.response).Return(nil)
	ExecuteRequest(suite.executor, suite.request, suite.response)

	suite.executor.AssertExpectations(suite.T())
	suite.Equal(suite.response.StatusCode(), fasthttp.StatusNotFound)
}

func (suite *ExecuteRequestTestSuite) TestExecuteRequestErr() {
	suite.executor.On("Do", suite.request, suite.response).Return(xerrors.New("Some error"))
	ExecuteRequest(suite.executor, suite.request, suite.response)

	suite.executor.AssertExpectations(suite.T())
	suite.Equal(suite.response.StatusCode(), fasthttp.StatusBadGateway)
}

func (suite *ExecuteRequestTestSuite) TestExecuteRequestTimeoutOK() {
	suite.executor.On("DoTimeout", suite.request, suite.response, time.Minute).Return(nil)
	ExecuteRequestTimeout(suite.executor, suite.request, suite.response, time.Minute)

	suite.executor.AssertExpectations(suite.T())
	suite.Equal(suite.response.StatusCode(), fasthttp.StatusNotFound)
}

func (suite *ExecuteRequestTestSuite) TestExecuteRequestTimeoutErr() {
	suite.executor.On("DoTimeout", suite.request, suite.response, time.Minute).Return(xerrors.New("Some error"))
	ExecuteRequestTimeout(suite.executor, suite.request, suite.response, time.Minute)

	suite.executor.AssertExpectations(suite.T())
	suite.Equal(suite.response.StatusCode(), fasthttp.StatusBadGateway)
}

func TestExecuteRequest(t *testing.T) {
	suite.Run(t, &ExecuteRequestTestSuite{})
}
