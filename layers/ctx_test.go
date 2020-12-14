package layers_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type EventChannelMock struct {
	mock.Mock
}

func (e *EventChannelMock) Send(ctx context.Context, eventType events.EventType, value interface{}, shardKey string) {
	e.Called(ctx, eventType, value, shardKey)
}

type ContextTestSuite struct {
	suite.Suite

	fhttpCtx      *fasthttp.RequestCtx
	eventsChannel *EventChannelMock
	ctx           *layers.Context
}

func (suite *ContextTestSuite) SetupTest() {
	suite.fhttpCtx = &fasthttp.RequestCtx{}

	remoteAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 65342,
	}

	suite.fhttpCtx.Init(&fasthttp.Request{}, remoteAddr, nil)

	suite.ctx = layers.AcquireContext()
	suite.eventsChannel = &EventChannelMock{}

	suite.ctx.Init(suite.fhttpCtx,
		"127.0.0.1:8000",
		suite.eventsChannel,
		"user",
		events.RequestTypeTLS)
}

func (suite *ContextTestSuite) TearDownTest() {
	suite.eventsChannel.AssertExpectations(suite.T())
	layers.ReleaseContext(suite.ctx)
}

func (suite *ContextTestSuite) TestEmptyContext() {
	ctx := layers.AcquireContext()
	defer layers.ReleaseContext(ctx)

	suite.Nil(ctx.Request())
	suite.Nil(ctx.Response())
	suite.Nil(ctx.RemoteAddr())
	suite.Nil(ctx.LocalAddr())

	ctx.Respond("msg", 100)
	ctx.Error(io.EOF)
	ctx.Hijack(nil, func(_, _ net.Conn) {})

	suite.False(ctx.Hijacked())

	ctx.Cancel()
	ctx.Deadline()

	suite.EqualError(ctx.Err(), "context canceled")

	ctx.Value(nil)

	suite.Nil(ctx.Get("a"))
	ctx.Delete("a")

	ctx.Set("a", "b")

	suite.Equal("b", ctx.Get("a"))

	ctx.Delete("a")

	suite.Nil(ctx.Get("a"))
}

func (suite *ContextTestSuite) TestRequest() {
	suite.Equal(&suite.fhttpCtx.Request, suite.ctx.Request())
}

func (suite *ContextTestSuite) TestResponse() {
	suite.Equal(&suite.fhttpCtx.Response, suite.ctx.Response())
}

func (suite *ContextTestSuite) TestRemoteAddr() {
	suite.Equal(suite.fhttpCtx.RemoteAddr(), suite.ctx.RemoteAddr())
}

func (suite *ContextTestSuite) TestLocalAddr() {
	suite.Equal(suite.fhttpCtx.LocalAddr(), suite.ctx.LocalAddr())
}

func (suite *ContextTestSuite) TestRespond() {
	suite.ctx.Respond("message", fasthttp.StatusBadGateway)

	resp := suite.ctx.Response()

	suite.Equal("message", string(resp.Body()))
	suite.Equal(fasthttp.StatusBadGateway, resp.StatusCode())
	suite.Equal("text/plain", string(resp.Header.ContentType()))
}

func (suite *ContextTestSuite) TestErrorGeneril() {
	suite.ctx.Error(io.EOF)

	resp := suite.ctx.Response()

	value := make(map[string]interface{})

	suite.NoError(json.Unmarshal(resp.Body(), &value))
	suite.Contains(string(resp.Body()), "EOF")
	suite.Equal(errors.DefaultChainStatusCode, resp.StatusCode())
	suite.Equal("application/json", string(resp.Header.ContentType()))
}

func (suite *ContextTestSuite) TestErrorCustom() {
	suite.ctx.Error(&errors.Error{
		Message:    "MY VALUE",
		StatusCode: fasthttp.StatusBadRequest,
		Err:        io.EOF,
	})

	resp := suite.ctx.Response()

	value := make(map[string]interface{})

	suite.NoError(json.Unmarshal(resp.Body(), &value))
	suite.Contains(string(resp.Body()), "EOF")
	suite.Contains(string(resp.Body()), "VALUE")
	suite.Equal(fasthttp.StatusBadRequest, resp.StatusCode())
	suite.Equal("application/json", string(resp.Header.ContentType()))
}

func (suite *ContextTestSuite) TestHijack() {
	suite.ctx.Hijack(nil, func(_, _ net.Conn) {})
	suite.True(suite.ctx.Hijacked())
}

func TestContext(t *testing.T) {
	suite.Run(t, &ContextTestSuite{})
}
