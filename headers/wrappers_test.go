package headers_test

import (
	"strings"
	"testing"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type RequestHeaderWrapperTestSuite struct {
	suite.Suite

	hdr *fasthttp.RequestHeader
	wrp headers.FastHTTPHeaderWrapper
}

func (suite *RequestHeaderWrapperTestSuite) SetupTest() {
	suite.hdr = &fasthttp.RequestHeader{}

	suite.hdr.SetMethod("GET")
	suite.hdr.SetHost("example.com")
	suite.hdr.SetRequestURI("http://example.com")
	suite.hdr.Set("Accept-Encoding", "deflate, gzip, br")

	suite.wrp = headers.NewRequestHeaderWrapper(suite.hdr)
}

func (suite *RequestHeaderWrapperTestSuite) TestCorrectRestore() {
	request := strings.Join([]string{
		"POST /lala HTTP/1.1",
		"Host: example.com",
		"accept: deflate",
	}, "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(request)))
	suite.Equal("GET", string(suite.hdr.Method()))
	suite.Equal("http://example.com", string(suite.hdr.RequestURI()))
	suite.Equal("example.com", string(suite.hdr.Host()))
}

func (suite *RequestHeaderWrapperTestSuite) TestDisableNormalizing() {
	suite.hdr.EnableNormalizing()

	request := strings.Join([]string{
		"GET / HTTP/1.1",
		"Host: example.com",
		"accept: deflate",
	}, "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(request)))
	suite.Contains(string(suite.hdr.Header()), "accept: ")
}

func (suite *RequestHeaderWrapperTestSuite) TestResetConnectionCloseNothing() {
	suite.hdr.SetConnectionClose()

	request := strings.Join([]string{
		"GET / HTTP/1.1",
		"Host: example.com",
		"accept: deflate",
	}, "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(request)))
	suite.False(suite.hdr.ConnectionClose())
}

func (suite *RequestHeaderWrapperTestSuite) TestResetConnectionCloseSet() {
	suite.hdr.SetConnectionClose()

	request := strings.Join([]string{
		"GET / HTTP/1.1",
		"Host: example.com",
		"accept: deflate",
		"connection: close",
	}, "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(request)))
	suite.True(suite.hdr.ConnectionClose())
}

func (suite *RequestHeaderWrapperTestSuite) TestRawHeaders() {
	suite.hdr.SetConnectionClose()

	request := []string{
		"Host: example.com",
		"accept: deflate",
		"connection: close",
	}
	fullRequest := strings.Join(append([]string{"GET / HTTP/1.1"}, request...), "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(fullRequest)))
	suite.Equal([]byte(strings.Join(request, "\r\n")+"\r\n\r\n"), suite.wrp.Headers())
}

type ResponseWrapperTestSuite struct {
	suite.Suite

	hdr *fasthttp.ResponseHeader
	wrp headers.FastHTTPHeaderWrapper
}

func (suite *ResponseWrapperTestSuite) SetupTest() {
	suite.hdr = &fasthttp.ResponseHeader{}

	suite.hdr.SetStatusCode(fasthttp.StatusCreated)
	suite.wrp = headers.NewResponseHeaderWrapper(suite.hdr)
}

func (suite *ResponseWrapperTestSuite) TestCorrectRestore() {
	request := strings.Join([]string{
		"HTTP/1.1 201 Created",
		"Host: example.com",
		"Content-encoding: deflate",
	}, "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(request)))
    suite.Equal(fasthttp.StatusCreated, suite.hdr.StatusCode())
}

func (suite *ResponseWrapperTestSuite) TestDisableNormalizing() {
	suite.hdr.EnableNormalizing()

	request := strings.Join([]string{
		"HTTP/1.1 201 Created",
		"Host: example.com",
		"Content-encoding: deflate",
	}, "\r\n") + "\r\n\r\n"

	suite.NoError(suite.wrp.Read(strings.NewReader(request)))
	suite.Contains(string(suite.hdr.Header()), "Content-encoding: ")
}

func TestRequestHeaderWrapper(t *testing.T) {
	suite.Run(t, &RequestHeaderWrapperTestSuite{})
}

func TestResponseHeaderWrapper(t *testing.T) {
    suite.Run(t, &ResponseWrapperTestSuite{})
}
