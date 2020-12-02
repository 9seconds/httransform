package events_test

import (
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/events"
	"github.com/stretchr/testify/suite"
)

type RequestTypeTestSuite struct {
	suite.Suite
}

func (suite *RequestTypeTestSuite) TestString() {
	r := events.RequestTypePlain

	suite.Contains(r.String(), "plain")
	suite.NotContains(r.String(), "tls")
	suite.NotContains(r.String(), "http")
	suite.NotContains(r.String(), "upgraded")

	r |= events.RequestTypeTLS

	suite.Contains(r.String(), "plain")
	suite.Contains(r.String(), "tls")
	suite.NotContains(r.String(), "http")
	suite.NotContains(r.String(), "upgraded")

	r |= events.RequestTypeHTTP

	suite.Contains(r.String(), "plain")
	suite.Contains(r.String(), "tls")
	suite.Contains(r.String(), "http")
	suite.NotContains(r.String(), "upgraded")

	r |= events.RequestTypeUpgraded

	suite.Contains(r.String(), "plain")
	suite.Contains(r.String(), "tls")
	suite.Contains(r.String(), "http")
	suite.Contains(r.String(), "upgraded")
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
		RequestType: events.RequestTypeHTTP | events.RequestTypeTLS,
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

func TestRequestType(t *testing.T) {
	suite.Run(t, &RequestTypeTestSuite{})
}

func TestRequestMeta(t *testing.T) {
	suite.Run(t, &RequestMetaTestSuite{})
}

func TestResponseMeta(t *testing.T) {
	suite.Run(t, &ResponseMetaTestSuite{})
}
