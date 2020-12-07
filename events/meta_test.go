package events_test

import (
	"errors"
	"io"
	"net"
	"os"
	"testing"

	"github.com/9seconds/httransform/v2/events"
	"github.com/stretchr/testify/suite"
)

type RequestTypeTestSuite struct {
	suite.Suite
}

func (suite *RequestTypeTestSuite) TestString() {
	suite.Contains(events.RequestTypeTunneled.String(), "tunneled:yes")
	suite.Contains(events.RequestTypeTLS.String(), "tls:yes")
	suite.Contains(events.RequestTypeUpgraded.String(), "upgraded:yes")
}

type RequestMetaTestSuite struct {
	suite.Suite
}

func (suite *RequestMetaTestSuite) TestString() {
	meta := events.RequestMeta{
		RequestID:   "reqid",
		User:        "user",
		Method:      "GET",
		Addr:        &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: 6000},
		RequestType: events.RequestTypeUpgraded,
	}

	suite.NoError(meta.URI.Parse(nil, []byte("http://example.com/image.gif")))

	value := meta.String()

	suite.Contains(value, "reqid")
	suite.Contains(value, "user")
	suite.Contains(value, "GET")
	suite.Contains(value, "127.0.0.1:6000")
	suite.Contains(value, "http://example.com/image.gif")
}

type ResponseMetaTestSuite struct {
	suite.Suite
}

func (suite *ResponseMetaTestSuite) TestString() {
	meta := events.ResponseMeta{
		RequestID:  "reqid",
		StatusCode: 404,
	}
	value := meta.String()

	suite.Contains(value, "reqid")
	suite.Contains(value, "404")
}

type ErrorMetaTestSuite struct {
	suite.Suite

	err *events.ErrorMeta
}

func (suite *ErrorMetaTestSuite) SetupTest() {
	err := &os.PathError{
		Op:  "TEST",
		Err: io.EOF,
	}
	suite.err = &events.ErrorMeta{
		RequestID: "reqid",
		Err:       err,
	}
}

func (suite *ErrorMetaTestSuite) TestInterface() {
	suite.Implements((*error)(nil), suite.err)
}

func (suite *ErrorMetaTestSuite) TestInterfaceIs() {
	suite.err.Err = io.EOF

	suite.True(errors.Is(suite.err, io.EOF))
}

func (suite *ErrorMetaTestSuite) TestInterfaceAs() {
	var err *os.PathError

	suite.True(errors.As(suite.err, &err))
	suite.Equal("TEST", err.Op)
}

func (suite *ErrorMetaTestSuite) TestError() {
	value := suite.err.Error()

	suite.Contains(value, "reqid")
	suite.Contains(value, "TEST")
}

type CommonErrorMetaTestSuite struct {
	suite.Suite

	err *events.CommonErrorMeta
}

func (suite *CommonErrorMetaTestSuite) SetupTest() {
	err := &os.PathError{
		Op:  "TEST",
		Err: io.EOF,
	}
	suite.err = &events.CommonErrorMeta{
		Method: "POST",
		Addr:   &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: 6001},
		Err:    err,
	}

	suite.NoError(suite.err.URI.Parse(nil, []byte("http://example.com/image.gif")))
}

func (suite *CommonErrorMetaTestSuite) TestInterface() {
	suite.Implements((*error)(nil), suite.err)
}

func (suite *CommonErrorMetaTestSuite) TestInterfaceIs() {
	suite.err.Err = io.EOF

	suite.True(errors.Is(suite.err, io.EOF))
}

func (suite *CommonErrorMetaTestSuite) TestInterfaceAs() {
	var err *os.PathError

	suite.True(errors.As(suite.err, &err))
	suite.Equal("TEST", err.Op)
}

func (suite *CommonErrorMetaTestSuite) TestError() {
	value := suite.err.Error()

	suite.Contains(value, "TEST")
	suite.Contains(value, "http://example.com/image.gif")
	suite.Contains(value, "POST")
	suite.Contains(value, "127.0.0.1:6001")
}

type TrafficMetaTestSuite struct {
	suite.Suite
}

func (suite *TrafficMetaTestSuite) TestString() {
	meta := events.TrafficMeta{
		ID:           "reqid",
		Addr:         &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: 6002},
		ReadBytes:    603,
		WrittenBytes: 500,
	}
	value := meta.String()

	suite.Contains(value, "reqid")
	suite.Contains(value, "127.0.0.1:6002")
	suite.Contains(value, "603")
	suite.Contains(value, "500")
}

func TestRequestType(t *testing.T) {
	suite.Run(t, &RequestTypeTestSuite{})
}

func TestRequestMeta(t *testing.T) {
	suite.Run(t, &RequestMetaTestSuite{})
}

func TestResponseMeta(t *testing.T) {
	suite.Run(t, &ResponseMetaTestSuite{})
}

func TestErrorMeta(t *testing.T) {
	suite.Run(t, &ErrorMetaTestSuite{})
}

func TestCommonErrorMeta(t *testing.T) {
	suite.Run(t, &CommonErrorMetaTestSuite{})
}

func TestTrafficMeta(t *testing.T) {
	suite.Run(t, &TrafficMetaTestSuite{})
}
