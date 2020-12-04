package conns_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/conns"
	"github.com/stretchr/testify/suite"
)

type FakeConn struct {
	rd *bytes.Buffer
}

func (f *FakeConn) Read(p []byte) (int, error)         { return f.rd.Read(p) }
func (f *FakeConn) Write(_ []byte) (int, error)        { return 0, io.EOF }
func (f *FakeConn) Close() error                       { return nil }
func (f *FakeConn) LocalAddr() net.Addr                { return nil }
func (f *FakeConn) RemoteAddr() net.Addr               { return nil }
func (f *FakeConn) SetDeadline(_ time.Time) error      { return nil }
func (f *FakeConn) SetReadDeadline(_ time.Time) error  { return nil }
func (f *FakeConn) SetWriteDeadline(_ time.Time) error { return nil }

type UnreadConnTestSuite struct {
	suite.Suite

	fconn *FakeConn
	uconn *conns.UnreadConn
}

func (suite *UnreadConnTestSuite) SetupTest() {
	suite.fconn = &FakeConn{
		rd: bytes.NewBufferString("1234567890"),
	}
	suite.uconn = conns.NewUnreadConn(suite.fconn)
}

func (suite *UnreadConnTestSuite) TestReadAll() {
	data, err := ioutil.ReadAll(suite.uconn)

	suite.NoError(err)
	suite.Equal("1234567890", string(data))

	p := make([]byte, 10)

	_, err = suite.uconn.Read(p)

	suite.Equal(err, io.EOF)
}

func (suite *UnreadConnTestSuite) TestUnread() {
	p := make([]byte, 5)

	_, err := io.ReadFull(suite.uconn, p)

	suite.Equal("12345", string(p))
	suite.NoError(err)

	suite.uconn.Unread()

	data, err := ioutil.ReadAll(suite.uconn)

	suite.NoError(err)
	suite.Equal("1234567890", string(data))
}

func (suite *UnreadConnTestSuite) TestSealed() {
	p := make([]byte, 5)

	_, err := io.ReadFull(suite.uconn, p)

	suite.Equal("12345", string(p))
	suite.NoError(err)

	suite.uconn.Seal()

	data, err := ioutil.ReadAll(suite.uconn)

	suite.NoError(err)
	suite.Equal("67890", string(data))
}

func TestUnreadConn(t *testing.T) {
	suite.Run(t, &UnreadConnTestSuite{})
}
