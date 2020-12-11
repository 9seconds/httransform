package layers_test

import (
	"context"
	"io"
	"net"
	"testing"

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

func TestContext(t *testing.T) {
	suite.Run(t, &ContextTestSuite{})
}
