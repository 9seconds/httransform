package upgrades_test

import (
	"bytes"
	"net"
	"time"

	"github.com/9seconds/httransform/v2/upgrades"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockConn struct {
	mock.Mock

	ReadBuffer  bytes.Buffer
	WriteBuffer bytes.Buffer
}

func (m *MockConn) Read(b []byte) (int, error) {
	m.Called(b)

	return m.ReadBuffer.Read(b)
}

func (m *MockConn) Write(b []byte) (int, error) {
	m.Called(b)

	return m.WriteBuffer.Write(b)
}

func (m *MockConn) Close() error {
	return m.Called().Error(0)
}

func (m *MockConn) LocalAddr() net.Addr {
	return m.Called().Get(0).(net.Addr)
}

func (m *MockConn) RemoteAddr() net.Addr {
	return m.Called().Get(0).(net.Addr)
}

func (m *MockConn) SetDeadline(tm time.Time) error {
	return m.Called(tm).Error(0)
}

func (m *MockConn) SetReadDeadline(tm time.Time) error {
	return m.Called(tm).Error(0)
}

func (m *MockConn) SetWriteDeadline(tm time.Time) error {
	return m.Called(tm).Error(0)
}

type UpgradesTestSuite struct {
	suite.Suite

	up         upgrades.Interface
	clientConn *MockConn
	netlocConn *MockConn
}

func (suite *UpgradesTestSuite) SetupTest() {
	suite.clientConn = &MockConn{}
	suite.netlocConn = &MockConn{}
}

func (suite *UpgradesTestSuite) TearDownTest() {
	suite.clientConn.AssertExpectations(suite.T())
	suite.netlocConn.AssertExpectations(suite.T())
}
