package http_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/http"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type MockRWC struct {
	mock.Mock
}

func (m *MockRWC) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockRWC) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockRWC) Close() error {
	return m.Called().Error(0)
}

type ExecuteTestSuite struct {
	suite.Suite

	rwc       *MockRWC
	ctx       context.Context
	ctxCancel context.CancelFunc
	req       *fasthttp.Request
	resp      *fasthttp.Response
}

func (suite *ExecuteTestSuite) SetupTest() {
	suite.rwc = &MockRWC{}
	suite.ctx, suite.ctxCancel = context.WithCancel(context.Background())
	suite.req = &fasthttp.Request{}
	suite.resp = &fasthttp.Response{}
}

func (suite *ExecuteTestSuite) TearDownTest() {
	suite.ctxCancel()
	suite.rwc.AssertExpectations(suite.T())
}

func (suite *ExecuteTestSuite) TestClosedContext() {
	suite.ctxCancel()
	suite.Error(http.Execute(suite.ctx, suite.rwc, suite.req, suite.resp))
}

func (suite *ExecuteTestSuite) TestBrokenWrite() {
	suite.req.SetRequestURI("http://example.com/path")
	suite.req.Header.Add("Accept-Encoding", "br")

	suite.rwc.On("Close").Maybe()
	suite.rwc.On("Write", mock.Anything).Return(10, io.EOF).Once()

	suite.Error(http.Execute(suite.ctx, suite.rwc, suite.req, suite.resp))
}

func (suite *ExecuteTestSuite) TestSentRequest() {
	suite.req.SetRequestURI("http://example.com/path")
	suite.req.Header.Add("Accept-Encoding", "br")

	reqLength := len([]byte(suite.req.String()))

	suite.rwc.On("Close").Maybe()
	suite.rwc.On("Write", mock.Anything).Return(reqLength, nil).Once().Run(func(args mock.Arguments) {
		sentRequest := string(args.Get(0).([]byte))
		reader := bufio.NewReader(strings.NewReader(sentRequest))
		req, err := nethttp.ReadRequest(reader)

		suite.NoError(err)
		suite.Equal("GET", req.Method)
		suite.Equal("/path", req.URL.Path)
		suite.Equal("example.com", req.Host)
		suite.Len(req.Header, 1)
		suite.Equal("br", req.Header.Get("Accept-Encoding"))
	})
	suite.rwc.On("Read", mock.Anything).Return(0, io.EOF)

	suite.Error(http.Execute(suite.ctx, suite.rwc, suite.req, suite.resp))
}

func (suite *ExecuteTestSuite) TestSentRequestContextClose() {
	suite.req.SetRequestURI("http://example.com/path")
	suite.req.Header.Add("Accept-Encoding", "br")

	reqLength := len([]byte(suite.req.String()))

	suite.rwc.On("Close").Return(nil).Once()
	suite.rwc.On("Write", mock.Anything).Return(reqLength, nil).Once().Run(func(args mock.Arguments) {
		suite.ctxCancel()
		time.Sleep(100 * time.Millisecond)
	})
	suite.rwc.On("Read", mock.Anything).Return(0, nil)

	suite.Error(http.Execute(suite.ctx, suite.rwc, suite.req, suite.resp))
	time.Sleep(100 * time.Millisecond)
}

func (suite *ExecuteTestSuite) TestResponseContentLength() {
	app := httpbin.NewHTTPBin()
	endpoint := httptest.NewServer(app.Handler())
	defer endpoint.Close()

	suite.req.SetRequestURI(endpoint.URL + "/ip")
	suite.req.Header.SetMethod("GET")

	addr := endpoint.Listener.Addr()

	conn, _ := net.Dial(addr.Network(), addr.String())
	defer conn.Close()

	v := map[string]interface{}{}

	suite.NoError(http.Execute(suite.ctx, conn, suite.req, suite.resp))
	suite.NoError(json.Unmarshal(suite.resp.Body(), &v))
	suite.Equal(nethttp.StatusOK, suite.resp.StatusCode())
}

func (suite *ExecuteTestSuite) TestResponseStream() {
	app := httpbin.NewHTTPBin()
	endpoint := httptest.NewServer(app.Handler())
	defer endpoint.Close()

	suite.req.SetRequestURI(endpoint.URL + "/stream-bytes/100")
	suite.req.Header.SetMethod("GET")

	addr := endpoint.Listener.Addr()

	conn, _ := net.Dial(addr.Network(), addr.String())
	defer conn.Close()

	suite.NoError(http.Execute(suite.ctx, conn, suite.req, suite.resp))
	suite.Equal(-1, suite.resp.Header.ContentLength())
	suite.Len(suite.resp.Body(), 100)
}

func (suite *ExecuteTestSuite) TestHEADRequest() {
	app := httpbin.NewHTTPBin()
	endpoint := httptest.NewServer(app.Handler())
	defer endpoint.Close()

	suite.req.SetRequestURI(endpoint.URL + "/stream-bytes/100")
	suite.req.Header.SetMethod("HEAD")

	addr := endpoint.Listener.Addr()

	conn, _ := net.Dial(addr.Network(), addr.String())
	defer conn.Close()

	suite.NoError(http.Execute(suite.ctx, conn, suite.req, suite.resp))
	suite.Equal(-2, suite.resp.Header.ContentLength())
	suite.Len(suite.resp.Body(), 0)
}

func TestExecute(t *testing.T) {
	suite.Run(t, &ExecuteTestSuite{})
}
