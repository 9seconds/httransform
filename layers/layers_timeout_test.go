package layers_test

import (
	"io"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/stretchr/testify/suite"
)

type LayerTimeoutTestSuite struct {
	BaseLayerTestSuite
}

func (suite *LayerTimeoutTestSuite) SetupTest() {
	suite.BaseLayerTestSuite.SetupTest()

	suite.l = layers.TimeoutLayer{
		Timeout: 100 * time.Millisecond,
	}
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

	time.Sleep(150 * time.Millisecond)

	suite.NoError(suite.l.OnResponse(suite.ctx, nil))

	select {
	case <-suite.ctx.Done():
	default:
		suite.FailNow("Layer has closed context unexpectedly")
	}
}

func (suite *LayerTimeoutTestSuite) TestSlowErr() {
	suite.NoError(suite.l.OnRequest(suite.ctx))

	time.Sleep(150 * time.Millisecond)

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
