package conns_test

import (
	"net"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/conns"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockedConn struct {
	mock.Mock
}

func (m *MockedConn) Read(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *MockedConn) Write(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *MockedConn) Close() error {
	return m.Called().Error(0)
}

func (m *MockedConn) LocalAddr() net.Addr {
	return m.Called().Get(0).(net.Addr)
}

func (m *MockedConn) RemoteAddr() net.Addr {
	return m.Called().Get(0).(net.Addr)
}

func (m *MockedConn) SetDeadline(tm time.Time) error {
	return m.Called(tm).Error(0)
}

func (m *MockedConn) SetReadDeadline(tm time.Time) error {
	return m.Called(tm).Error(0)
}

func (m *MockedConn) SetWriteDeadline(tm time.Time) error {
	return m.Called(tm).Error(0)
}

type UnsafeMockedConn struct {
	MockedConn
}

func (u *UnsafeMockedConn) UnsafeConn() net.Conn {
	return u.Called().Get(0).(net.Conn)
}

type UnhijackTestSuite struct {
	suite.Suite

	raw     *MockedConn
	wrapped *UnsafeMockedConn
}

func (suite *UnhijackTestSuite) SetupTest() {
	suite.raw = &MockedConn{}
	suite.wrapped = &UnsafeMockedConn{}

	suite.wrapped.On("UnsafeConn").Return(suite.raw)
}

func (suite *UnhijackTestSuite) TearDownTest() {
	suite.raw.AssertExpectations(suite.T())
	suite.wrapped.AssertExpectations(suite.T())
}

func (suite *UnhijackTestSuite) TestKeepOpened() {
	conns.FixHijackHandler(func(net.Conn) bool {
		return false
	})(suite.wrapped)
}

func (suite *UnhijackTestSuite) TestClose() {
	suite.raw.On("Close").Return(nil)

	conns.FixHijackHandler(func(net.Conn) bool {
		return true
	})(suite.wrapped)
}

func TestUnhijack(t *testing.T) {
	suite.Run(t, &UnhijackTestSuite{})
}
