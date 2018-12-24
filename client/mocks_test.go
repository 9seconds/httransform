package client

import (
	"io"
	"net"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockedConn struct {
	mock.Mock
}

func (m *MockedConn) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockedConn) Write(p []byte) (int, error) {
	args := m.Called(p)
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

func (m *MockedConn) SetDeadline(t time.Time) error {
	return m.Called(t).Error(0)
}

func (m *MockedConn) SetReadDeadline(t time.Time) error {
	return m.Called(t).Error(0)
}

func (m *MockedConn) SetWriteDeadline(t time.Time) error {
	return m.Called(t).Error(0)
}

type MockedConnReader struct {
	MockedConn

	reader io.Reader
}

func (m *MockedConnReader) Read(b []byte) (int, error) {
	return m.reader.Read(b)
}

type MockedBaseDialer struct {
	mock.Mock
}

func (m *MockedBaseDialer) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	args := m.Called(addr, timeout)
	return args.Get(0).(net.Conn), args.Error(1)
}

type MockedDialer struct {
	mock.Mock
}

func (m *MockedDialer) Dial(addr string) (net.Conn, error) {
	args := m.Called(addr)
	return args.Get(0).(net.Conn), args.Error(1)
}

func (m *MockedDialer) Release(conn net.Conn, addr string) {
	m.Called(conn, addr)
}

func (m *MockedDialer) NotifyClosed(addr string) {
	m.Called(addr)
}
