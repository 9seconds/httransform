package events_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/events"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProcessorMock struct {
	mock.Mock
}

func (p *ProcessorMock) Process(evt events.Event) {
	p.Called(evt)
}

func (p *ProcessorMock) Shutdown() {
	p.Called()
}

type ProcessorMockFactory struct {
	All []*ProcessorMock
}

func (p *ProcessorMockFactory) Make() events.Processor {
	newProc := &ProcessorMock{}

	p.All = append(p.All, newProc)

	return newProc
}

func (p *ProcessorMockFactory) CalledMocks() []*ProcessorMock {
	rv := []*ProcessorMock{}

	for _, v := range p.All {
		if len(v.Calls) > 0 {
			rv = append(rv, v)
		}
	}

	return rv
}

func (p *ProcessorMockFactory) Assert(t *testing.T) {
	for _, v := range p.CalledMocks() {
		v.AssertExpectations(t)
	}
}

type ChannelTestSuite struct {
	suite.Suite

	factory *ProcessorMockFactory
	cancel  context.CancelFunc
	channel *events.Channel
}

func (suite *ChannelTestSuite) SetupTest() {
	ctx, cancel := context.WithCancel(context.Background())
	suite.cancel = cancel
	suite.factory = &ProcessorMockFactory{
		All: []*ProcessorMock{},
	}
	suite.channel = events.NewChannel(ctx, suite.factory.Make)

	sleep()
}

func (suite *ChannelTestSuite) TearDownTest() {
	suite.cancel()

	sleep()

	suite.factory.Assert(suite.T())
}

func (suite *ChannelTestSuite) TestDefaultSharding() {
	if runtime.NumCPU() == 1 {
		return
	}

	for _, v := range suite.factory.All {
		v.On("Process", mock.Anything).Maybe()
		v.On("Shutdown").Once()
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		suite.channel.Send(ctx, events.EventTypeStartRequest, nil, "")
	}

	sleep()

	suite.Greater(len(suite.factory.CalledMocks()), 1)
}

func (suite *ChannelTestSuite) TestRouteToTheSameShard() {
	for _, v := range suite.factory.All {
		v.On("Process", mock.Anything).Maybe()
		v.On("Shutdown").Once()
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		suite.channel.Send(ctx, events.EventTypeStartRequest, nil, "hello")
	}

	sleep()

	suite.Len(suite.factory.CalledMocks(), 1)

	mocked := suite.factory.CalledMocks()[0]

	mocked.AssertNumberOfCalls(suite.T(), "Process", 10)
}

func (suite *ChannelTestSuite) TestProcessEvent() {
	for _, v := range suite.factory.All {
		v.On("Shutdown").Once()
		v.On("Process", mock.Anything).Maybe().Run(func(args mock.Arguments) {
			evt := args.Get(0).(events.Event)

			suite.Equal(events.EventTypeNewCertificate, evt.Type)
			suite.Equal(111, evt.Value)
			suite.WithinDuration(time.Now(), evt.Time, time.Second)
		})
	}

	ctx := context.Background()

	suite.channel.Send(ctx, events.EventTypeNewCertificate, 111, "hello")

	sleep()

	suite.Len(suite.factory.CalledMocks(), 1)
}

func TestChannel(t *testing.T) {
	suite.Run(t, &ChannelTestSuite{})
}

func sleep() {
	time.Sleep(20 * time.Millisecond)
}
