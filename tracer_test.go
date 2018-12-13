package httransform

import (
	"testing"
	"time"

	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type NoopTracerTestSuite struct {
	suite.Suite
}

func (suite *NoopTracerTestSuite) TestDummy() {
	tr := &NoopTracer{}
	tr.StartOnRequest()
	tr.StartOnResponse()
	tr.StartExecute()
	tr.FinishOnRequest(errors.New("Error"))
	tr.FinishOnResponse()
	tr.FinishExecute()
	tr.Clear()
	tr.Dump(nil, nil)
}

type LogTracerTestSuite struct {
	suite.Suite

	pool   *TracerPool
	tracer *LogTracer
}

func (suite *LogTracerTestSuite) SetupSuite() {
	suite.pool = NewTracerPool(func() Tracer {
		return &LogTracer{}
	})
}

func (suite *LogTracerTestSuite) SetupTest() {
	suite.tracer = suite.pool.acquire().(*LogTracer)
}

func (suite *LogTracerTestSuite) TearDownTest() {
	suite.pool.release(suite.tracer)
}

func (suite *LogTracerTestSuite) TestStartOnRequestTime() {
	suite.True(suite.tracer.startOnRequestTime.IsZero())
	suite.tracer.StartOnRequest()
	suite.WithinDuration(suite.tracer.startOnRequestTime, time.Now(), time.Second)
}

func (suite *LogTracerTestSuite) TestStartOnRequestDoubleCall() {
	suite.tracer.StartOnRequest()
	suite.Panics(suite.tracer.StartOnRequest)
}

func (suite *LogTracerTestSuite) TestStartOnResponseTime() {
	suite.True(suite.tracer.startOnResponseTime.IsZero())
	suite.tracer.StartOnResponse()
	suite.WithinDuration(suite.tracer.startOnResponseTime, time.Now(), time.Second)
}

func (suite *LogTracerTestSuite) TestStartOnResponseDoubleCall() {
	suite.tracer.StartOnResponse()
	suite.Panics(suite.tracer.StartOnResponse)
}

func (suite *LogTracerTestSuite) TestStartExecute() {
	suite.True(suite.tracer.startExecuteTime.IsZero())
	suite.tracer.StartExecute()
	suite.WithinDuration(suite.tracer.startExecuteTime, time.Now(), time.Second)
}

func (suite *LogTracerTestSuite) TestStartExecuteDoubleCall() {
	suite.tracer.StartExecute()
	suite.Panics(suite.tracer.StartExecute)
}

func (suite *LogTracerTestSuite) TestFinishOnRequestTime() {
	suite.tracer.StartOnRequest()
	suite.tracer.FinishOnRequest(nil)
	suite.Len(suite.tracer.onRequestDurations, 1)
	suite.True(suite.tracer.onRequestDurations[0] < time.Second)
	suite.True(suite.tracer.startOnRequestTime.IsZero())
}

func (suite *LogTracerTestSuite) TestFinishOnRequestEager() {
	suite.Panics(func() { suite.tracer.FinishOnRequest(nil) })
}

func (suite *LogTracerTestSuite) TestFinishOnRequestDoubleCall() {
	suite.tracer.StartOnRequest()
	suite.tracer.FinishOnRequest(nil)
	suite.Panics(func() { suite.tracer.FinishOnRequest(nil) })
}

func (suite *LogTracerTestSuite) TestFinishOnResponseTime() {
	suite.tracer.StartOnResponse()
	suite.tracer.FinishOnResponse()
	suite.Len(suite.tracer.onResponseDurations, 1)
	suite.True(suite.tracer.onResponseDurations[0] < time.Second)
	suite.True(suite.tracer.startOnResponseTime.IsZero())
}

func (suite *LogTracerTestSuite) TestFinishOnResponseEager() {
	suite.Panics(suite.tracer.FinishOnResponse)
}

func (suite *LogTracerTestSuite) TestFinishOnResponseDoubleCall() {
	suite.tracer.StartOnResponse()
	suite.tracer.FinishOnResponse()
	suite.Panics(suite.tracer.FinishOnResponse)
}

func (suite *LogTracerTestSuite) TestFinishExecuteTime() {
	suite.tracer.StartExecute()
	suite.tracer.FinishExecute()
	suite.True(suite.tracer.executeDuration < time.Second)
	suite.True(suite.tracer.startExecuteTime.IsZero())
}

func (suite *LogTracerTestSuite) TestFinishExecuteEager() {
	suite.Panics(suite.tracer.FinishExecute)
}

func (suite *LogTracerTestSuite) TestFinishExecuteDoubleCall() {
	suite.tracer.StartExecute()
	suite.tracer.FinishExecute()
	suite.Panics(suite.tracer.FinishExecute)
}

func (suite *LogTracerTestSuite) TestClear() {
	suite.tracer.StartOnRequest()
	suite.tracer.FinishOnRequest(nil)
	suite.tracer.StartExecute()
	suite.tracer.FinishExecute()
	suite.tracer.StartOnResponse()
	suite.tracer.FinishOnResponse()
	suite.tracer.Clear()

	suite.True(suite.tracer.startOnRequestTime.IsZero())
	suite.True(suite.tracer.startOnResponseTime.IsZero())
	suite.True(suite.tracer.startExecuteTime.IsZero())
	suite.Equal(suite.tracer.executeDuration, time.Duration(0))
	suite.Len(suite.tracer.onRequestDurations, 0)
	suite.Len(suite.tracer.onResponseDurations, 0)
}

func (suite *LogTracerTestSuite) TestDump() {
	suite.tracer.StartOnRequest()
	suite.tracer.FinishOnRequest(nil)
	suite.tracer.StartExecute()
	suite.tracer.FinishExecute()
	suite.tracer.StartOnResponse()
	suite.tracer.FinishOnResponse()
	suite.tracer.Clear()

	state := getLayerState()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	ctx := &fasthttp.RequestCtx{
		Request:  *req,  // nolint: govet
		Response: *resp, // nolint: govet
	}
	requestHeaderSet := getHeaderSet()
	responseHeaderSet := getHeaderSet()
	user := []byte("user")
	password := []byte("password")
	initLayerState(state, ctx, requestHeaderSet, responseHeaderSet, true, user, password)

	suite.tracer.Dump(state, &NoopLogger{})
}

func TestNoopTracer(t *testing.T) {
	suite.Run(t, &NoopTracerTestSuite{})
}

func TestLogTracer(t *testing.T) {
	suite.Run(t, &LogTracerTestSuite{})
}
