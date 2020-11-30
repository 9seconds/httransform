package conns_test

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/9seconds/httransform/v2/conns"
	"github.com/9seconds/httransform/v2/events"
	"github.com/stretchr/testify/suite"
)

type TraffcConnTestSuite struct {
	suite.Suite

	cancel  context.CancelFunc
	channel chan *events.Event
	addr    *net.TCPAddr
	raw     *MockedConn
	conn    *conns.TrafficConn
}

func (suite *TraffcConnTestSuite) SetupTest() {
	ctx, cancel := context.WithCancel(context.Background())

	suite.cancel = cancel
	suite.channel = make(chan *events.Event, 1)
	suite.raw = &MockedConn{}
	suite.addr = &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 1000,
	}
	suite.conn = &conns.TrafficConn{
		Conn:    suite.raw,
		Context: ctx,
		ID:      "HELLO",
		Events:  suite.channel,
	}
}

func (suite *TraffcConnTestSuite) TearDownTest() {
	suite.cancel()
	suite.raw.AssertExpectations(suite.T())
}

func (suite *TraffcConnTestSuite) TestReadOk() {
	suite.raw.On("Read", []byte(nil)).Once().Return(10, nil)
	suite.raw.On("RemoteAddr").Return(suite.addr)
	suite.raw.On("Close").Return(nil)

	n, err := suite.conn.Read(nil)

	suite.Equal(10, n)
	suite.NoError(err)

	suite.NoError(suite.conn.Close())

	var meta *events.TrafficMeta

	select {
	case evt := <-suite.channel:
		meta = evt.Value().(*events.TrafficMeta)
	default:
		suite.FailNow("Haven't found an event")
	}

	suite.Equal(suite.conn.ID, meta.ID)
	suite.Equal(suite.addr, meta.Addr)
	suite.EqualValues(10, meta.ReadBytes)
	suite.EqualValues(0, meta.WrittenBytes)
}

func (suite *TraffcConnTestSuite) TestReadErr() {
	suite.raw.On("Read", []byte(nil)).Once().Return(10, io.EOF)
	suite.raw.On("RemoteAddr").Return(suite.addr)
	suite.raw.On("Close").Return(nil)

	n, err := suite.conn.Read(nil)

	suite.Equal(10, n)
	suite.Error(io.EOF, err)

	suite.NoError(suite.conn.Close())

	var meta *events.TrafficMeta

	select {
	case evt := <-suite.channel:
		meta = evt.Value().(*events.TrafficMeta)
	default:
		suite.FailNow("Haven't found an event")
	}

	suite.Equal(suite.conn.ID, meta.ID)
	suite.Equal(suite.addr, meta.Addr)
	suite.EqualValues(10, meta.ReadBytes)
	suite.EqualValues(0, meta.WrittenBytes)
}

func (suite *TraffcConnTestSuite) TestWriteOk() {
	suite.raw.On("Write", []byte(nil)).Once().Return(10, nil)
	suite.raw.On("RemoteAddr").Return(suite.addr)
	suite.raw.On("Close").Return(nil)

	n, err := suite.conn.Write(nil)

	suite.Equal(10, n)
	suite.NoError(err)

	suite.NoError(suite.conn.Close())

	var meta *events.TrafficMeta

	select {
	case evt := <-suite.channel:
		meta = evt.Value().(*events.TrafficMeta)
	default:
		suite.FailNow("Haven't found an event")
	}

	suite.Equal(suite.conn.ID, meta.ID)
	suite.Equal(suite.addr, meta.Addr)
	suite.EqualValues(0, meta.ReadBytes)
	suite.EqualValues(10, meta.WrittenBytes)
}

func (suite *TraffcConnTestSuite) TestWriteErr() {
	suite.raw.On("Write", []byte(nil)).Once().Return(10, io.EOF)
	suite.raw.On("RemoteAddr").Return(suite.addr)
	suite.raw.On("Close").Return(nil)

	n, err := suite.conn.Write(nil)

	suite.Equal(10, n)
	suite.Error(io.EOF, err)

	suite.NoError(suite.conn.Close())

	var meta *events.TrafficMeta

	select {
	case evt := <-suite.channel:
		meta = evt.Value().(*events.TrafficMeta)
	default:
		suite.FailNow("Haven't found an event")
	}

	suite.Equal(suite.conn.ID, meta.ID)
	suite.Equal(suite.addr, meta.Addr)
	suite.EqualValues(0, meta.ReadBytes)
	suite.EqualValues(10, meta.WrittenBytes)
}

func (suite *TraffcConnTestSuite) TestMixed() {
	suite.raw.On("Write", []byte(nil)).Twice().Return(10, nil)
	suite.raw.On("Read", []byte(nil)).Once().Return(5, nil)
	suite.raw.On("RemoteAddr").Return(suite.addr)
	suite.raw.On("Close").Return(nil)

	_, err := suite.conn.Write(nil)

    suite.NoError(err)

	_, err = suite.conn.Read(nil)

    suite.NoError(err)

	_, err = suite.conn.Write(nil)

    suite.NoError(err)
    suite.NoError(suite.conn.Close())
    suite.NoError(suite.conn.Close())

	var meta *events.TrafficMeta

	select {
	case evt := <-suite.channel:
		meta = evt.Value().(*events.TrafficMeta)
	default:
		suite.FailNow("Haven't found an event")
	}

	suite.EqualValues(5, meta.ReadBytes)
	suite.EqualValues(20, meta.WrittenBytes)

    suite.Empty(suite.channel)
}

func TestTrafficConn(t *testing.T) {
	suite.Run(t, &TraffcConnTestSuite{})
}
