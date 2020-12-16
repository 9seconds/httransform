package headers_test

import (
	"strings"
	"testing"

	"github.com/9seconds/httransform/v2/headers"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type HeadersTestSuite struct {
	suite.Suite

	original *fasthttp.RequestHeader
	hdrs     *headers.Headers
}

func (suite *HeadersTestSuite) SetupTest() {
	suite.hdrs = headers.AcquireHeaderSet()
	req := strings.Join([]string{
		"GET / HTTP/1.1",
		"Host: example.com",
		"Accept-Encoding: deflate, gzip, br",
		"Accept-Language: en-US",
	}, "\r\n") + "\r\n\r\n"
	suite.original = &fasthttp.RequestHeader{}
	wrapper := headers.NewRequestHeaderWrapper(suite.original)

	suite.NoError(wrapper.Read(strings.NewReader(req)))
	suite.NoError(suite.hdrs.Init(wrapper))
}

func (suite *HeadersTestSuite) TestCheckHeaders() {
	suite.Len(suite.hdrs.Headers, 2)

	headerNames := map[string]bool{
		"Accept-Encoding": true,
		"Accept-Language": true,
	}

	for _, v := range suite.hdrs.Headers {
		delete(headerNames, v.Name())
	}

	suite.Empty(headerNames)
}

func (suite *HeadersTestSuite) TestGetAll() {
	suite.hdrs.Append("accept-encoding", "hello")

	insensitive := suite.hdrs.GetAll("ACCEPT-ENCODING")

	suite.Len(insensitive, 2)
	suite.Equal("Accept-Encoding", insensitive[0].Name())
	suite.Equal("deflate, gzip, br", insensitive[0].Value())
	suite.Equal("accept-encoding", insensitive[1].Name())
	suite.Equal("hello", insensitive[1].Value())
}

func (suite *HeadersTestSuite) TestGetAllNothing() {
	suite.Empty(suite.hdrs.GetAll("HELLO"))
}

func (suite *HeadersTestSuite) TestGetAllExact() {
	suite.hdrs.Append("accept-encoding", "hello")

	insensitive := suite.hdrs.GetAllExact("accept-encoding")

	suite.Len(insensitive, 1)
	suite.Equal("accept-encoding", insensitive[0].Name())
	suite.Equal("hello", insensitive[0].Value())
}

func (suite *HeadersTestSuite) TestGetAllExactNothing() {
	suite.Empty(suite.hdrs.GetAllExact("HELLO"))
}

func (suite *HeadersTestSuite) TestGetLast() {
	suite.hdrs.Append("accept-encoding", "hello")

	header := suite.hdrs.GetLast("Accept-encoding")

	suite.Equal("accept-encoding", header.Name())
	suite.Equal("hello", header.Value())
}

func (suite *HeadersTestSuite) TestGetLastNothing() {
	suite.Nil(suite.hdrs.GetLast("Accept"))
}

func (suite *HeadersTestSuite) TestGetLastExact() {
	suite.hdrs.Append("Accept-Encoding", "qq")
	suite.hdrs.Append("accept-encoding", "hello")

	header := suite.hdrs.GetLastExact("Accept-Encoding")

	suite.Equal("Accept-Encoding", header.Name())
	suite.Equal("qq", header.Value())
}

func (suite *HeadersTestSuite) TestGetLastExactNothing() {
	suite.Nil(suite.hdrs.GetLastExact("HELLO"))
}

func (suite *HeadersTestSuite) TestGetFirst() {
	suite.hdrs.Append("accept-encoding", "hello")

	header := suite.hdrs.GetFirst("Accept-encoding")

	suite.Equal("Accept-Encoding", header.Name())
	suite.Equal("deflate, gzip, br", header.Value())
}

func (suite *HeadersTestSuite) TestGetFirstNothing() {
	suite.Nil(suite.hdrs.GetFirst("HELLO"))
}

func (suite *HeadersTestSuite) TestGetFirstExact() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.Append("accept-encoding", "hello2")

	header := suite.hdrs.GetFirstExact("accept-encoding")

	suite.Equal("accept-encoding", header.Name())
	suite.Equal("hello", header.Value())
}

func (suite *HeadersTestSuite) TestSetNoCleanup() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.Set("accept-encoding", "NewValue", false)

	header := suite.hdrs.GetFirst("accept-encoding")

	suite.Equal("NewValue", header.Value())

	header = suite.hdrs.GetLast("accept-encoding")

	suite.Equal("hello", header.Value())
}

func (suite *HeadersTestSuite) TestSetUnknown() {
	suite.hdrs.Set("hello", "NewValue", false)
	suite.Len(suite.hdrs.Headers, 3)

	header := suite.hdrs.GetFirst("hello")

	suite.Equal("NewValue", header.Value())
}

func (suite *HeadersTestSuite) TestSetCleanup() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.Set("accept-encoding", "NewValue", true)

	headers := suite.hdrs.GetAll("accept-encoding")

	suite.Len(headers, 1)
	suite.Equal("Accept-Encoding", headers[0].Name())
	suite.Equal("NewValue", headers[0].Value())
}

func (suite *HeadersTestSuite) TestSetExactNoCleanup() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.SetExact("accept-encoding", "NewValue", false)

	header := suite.hdrs.GetFirst("accept-encoding")

	suite.Equal("deflate, gzip, br", header.Value())

	header = suite.hdrs.GetLast("accept-encoding")

	suite.Equal("NewValue", header.Value())
}

func (suite *HeadersTestSuite) TestSetExactUnknown() {
	suite.hdrs.SetExact("hello", "NewValue", false)

	header := suite.hdrs.GetFirst("hello")

	suite.Equal("hello", header.Name())
	suite.Equal("NewValue", header.Value())
}

func (suite *HeadersTestSuite) TestSetExactCleanup() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.Append("accept-encoding", "hello2")
	suite.hdrs.SetExact("accept-encoding", "NewValue", true)

	suite.Len(suite.hdrs.GetAll("accept-encoding"), 2)

	headers := suite.hdrs.GetAllExact("accept-encoding")

	suite.Len(headers, 1)
	suite.Equal("NewValue", headers[0].Value())
}

func (suite *HeadersTestSuite) TestRemove() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.Append("accept-encoding", "hello2")

	suite.hdrs.Remove("accept-encoding")

	suite.Empty(suite.hdrs.GetAll("accept-encoding"))
}

func (suite *HeadersTestSuite) TestRemoveExact() {
	suite.hdrs.Append("accept-encoding", "hello")
	suite.hdrs.Append("accept-encoding", "hello2")

	suite.hdrs.RemoveExact("accept-encoding")

	suite.Empty(suite.hdrs.GetAllExact("accept-encoding"))
	suite.Len(suite.hdrs.GetAll("accept-encoding"), 1)
}

func (suite *HeadersTestSuite) TestSync() {
	suite.hdrs.Append("accept-encoding", "hello")

	suite.NoError(suite.hdrs.Sync())
    suite.Contains(string(suite.original.RawHeaders()), "accept-encoding: ")
}

func (suite *HeadersTestSuite) TearDownTest() {
	headers.ReleaseHeaderSet(suite.hdrs)
}

func TestHeaders(t *testing.T) {
	suite.Run(t, &HeadersTestSuite{})
}
