package layers_test

import (
	"context"
	"net"
	"time"

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

type BaseLayerTestSuite struct {
	suite.Suite

	l             layers.Layer
	ctx           *layers.Context
	eventsChannel *EventChannelMock
}

func (suite *BaseLayerTestSuite) SetupTest() {
	fhttpCtx := &fasthttp.RequestCtx{}

	remoteAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 65342,
	}

	fhttpCtx.Init(&fasthttp.Request{}, remoteAddr, nil)

	suite.ctx = layers.AcquireContext()
	suite.eventsChannel = &EventChannelMock{}

	suite.ctx.Init(fhttpCtx,
		"127.0.0.1:8000",
		suite.eventsChannel,
		"user",
		events.RequestTypeTLS)

	suite.l = layers.TimeoutLayer{
		Timeout: 50 * time.Millisecond,
	}
}

func (suite *BaseLayerTestSuite) TearDownTest() {
	layers.ReleaseContext(suite.ctx)
}
