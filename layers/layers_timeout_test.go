package layers_test

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type LayerTimeoutTestSuite struct {
	suite.Suite

	l             layers.Layer
	ctx           *layers.Context
	eventsChannel *EventChannelMock
}

func (suite *LayerTimeoutTestSuite) SetupTest() {
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

func (suite *LayerTimeoutTestSuite) TearDownTest() {
	layers.ReleaseContext(suite.ctx)
}

func (suite *LayerTimeoutTestSuite) TestFastOk() {
	suite.NoError(suite.l.OnRequest(suite.ctx))

	time.Sleep(20 * time.Millisecond)

	suite.NoError(suite.l.OnResponse(suite.ctx, nil))

	select {
	case <-suite.ctx.Done():
		suite.FailNow("Layer has closed context unexpectedly")
	default:
	}
}

func (suite *LayerTimeoutTestSuite) TestFastErr() {
	suite.NoError(suite.l.OnRequest(suite.ctx))

	time.Sleep(20 * time.Millisecond)

	suite.Equal(io.EOF, suite.l.OnResponse(suite.ctx, io.EOF))

	select {
	case <-suite.ctx.Done():
		suite.FailNow("Layer has closed context unexpectedly")
	default:
	}
}

func (suite *LayerTimeoutTestSuite) TestSlowOk() {
	suite.NoError(suite.l.OnRequest(suite.ctx))

	time.Sleep(80 * time.Millisecond)

	suite.NoError(suite.l.OnResponse(suite.ctx, nil))

	select {
	case <-suite.ctx.Done():
	default:
		suite.FailNow("Layer has closed context unexpectedly")
	}
}

func (suite *LayerTimeoutTestSuite) TestSlowErr() {
	suite.NoError(suite.l.OnRequest(suite.ctx))

	time.Sleep(80 * time.Millisecond)

	suite.Equal(io.EOF, suite.l.OnResponse(suite.ctx, io.EOF))

	select {
	case <-suite.ctx.Done():
	default:
		suite.FailNow("Layer has closed context unexpectedly")
	}
}

func TestLayerTimeout(t *testing.T) {
	suite.Run(t, &LayerTimeoutTestSuite{})
}
