package client

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
)

type ConnsTestSuite struct {
	suite.Suite

	c      *conns
	dialer *MockedDialer
	addr   string
}

func (suite *ConnsTestSuite) SetupTest() {
	suite.addr = "myaddr:8080"
	suite.dialer = &MockedDialer{}
	suite.c, _ = newConns(suite.addr, true, 2,
		&tls.Config{InsecureSkipVerify: true}, suite.dialer.Dial)
	go suite.c.run()
}

func (suite *ConnsTestSuite) TestGet() {
	conn := &MockedConn{}
	suite.dialer.On("Dial", suite.addr).Return(conn, nil)

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
	suite.dialer.On("Dial", suite.addr).Return(conn, nil)

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
	suite.dialer.On("Dial", suite.addr).Return(conn, nil)

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
	suite.dialer.On("Dial", suite.addr).Return(conn, nil)

	gotConn, _ := suite.c.get(time.Millisecond)
	suite.c.put(gotConn)

	time.Sleep((connsGCAfter + 1) * connsGCTick)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestStopped() {
	conn := &MockedConn{}
	conn.On("Close").Return(nil)
	suite.dialer.On("Dial", suite.addr).Return(conn, nil)

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
	suite.dialer.On("Dial", suite.addr).Return(conn, errors.New("error"))

	_, err := suite.c.get(time.Millisecond)
	suite.NotNil(err)

	conn.AssertExpectations(suite.T())
	suite.dialer.AssertExpectations(suite.T())
}

func (suite *ConnsTestSuite) TestNegativeFreeSlot() {
	_, err := newConns(suite.addr, true, 0,
		&tls.Config{InsecureSkipVerify: true}, suite.dialer.Dial)
	suite.NotNil(err)

	_, err = newConns(suite.addr, true, -1,
		&tls.Config{InsecureSkipVerify: true}, suite.dialer.Dial)
	suite.NotNil(err)
}

func TestConns(t *testing.T) {
	suite.Run(t, &ConnsTestSuite{})
}
