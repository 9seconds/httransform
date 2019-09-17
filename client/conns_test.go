package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type ConnsTestSuite struct {
	suite.Suite

	c            *conns
	dialer       *MockedBaseDialer
	addr         string
	obsoleteChan chan string
}

func (suite *ConnsTestSuite) SetupTest() {
	suite.obsoleteChan = make(chan string)
	suite.addr = "myaddr:8080"
	suite.dialer = &MockedBaseDialer{}
	suite.c = newConns(suite.addr, suite.dialer.Dial, time.Second, 2, suite.obsoleteChan)
	suite.c.use()
	go suite.c.run()
}

func (suite *ConnsTestSuite) TestGet() {
	conn := &MockedConn{}
	suite.dialer.On("Dial", suite.addr, time.Second).Return(conn, nil)

	_, err := suite.c.get(time.Millisecond)
	suite.Nil(err)

	_, err = suite.c.get(time.Millisecond)
	suite.Nil(err)

	_, err = suite.c.get(time.Millisecond)
	suite.NotNil(err)

	suite.False(suite.c.isObsolete())

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestPut() {
	conn := &MockedConn{}
	suite.dialer.On("Dial", suite.addr, time.Second).Return(conn, nil)

	_, _ = suite.c.get(time.Millisecond)
	gotConn, _ := suite.c.get(time.Millisecond)
	suite.c.put(gotConn)

	_, err := suite.c.get(time.Millisecond)
	suite.Nil(err)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestClosed() {
	conn := &MockedConn{}
	suite.dialer.On("Dial", suite.addr, time.Second).Return(conn, nil)

	_, _ = suite.c.get(time.Millisecond)
	_, _ = suite.c.get(time.Millisecond)
	suite.c.notifyClosed()

	_, err := suite.c.get(time.Millisecond)
	suite.Nil(err)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestGC() {
	conn := &MockedConn{}
	conn.On("Close").Return(nil)
	suite.dialer.On("Dial", suite.addr, time.Second).Return(conn, nil)

	gotConn, _ := suite.c.get(time.Millisecond)
	suite.c.put(gotConn)

	time.Sleep((connsGCAfter + 1) * connsGCTick)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestStopped() {
	conn := &MockedConn{}
	conn.On("Close").Return(nil)
	suite.dialer.On("Dial", suite.addr, time.Second).Return(conn, nil)

	_, _ = suite.c.get(time.Millisecond)
	gotConn, _ := suite.c.get(time.Millisecond)
	suite.c.put(gotConn)
	suite.c.stop()
	suite.c.stop()
	suite.c.put(gotConn)
	suite.c.notifyClosed()

	_, err := suite.c.get(time.Millisecond)
	suite.NotNil(err)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestGetError() {
	conn := &MockedConn{}
	conn.On("Close").Return(nil)
	suite.dialer.On("Dial", suite.addr, time.Second).Return(conn, xerrors.New("error"))

	_, err := suite.c.get(time.Millisecond)
	suite.NotNil(err)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func TestConns(t *testing.T) {
	suite.Run(t, &ConnsTestSuite{})
}
