package client

import (
	"io"
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
)

type BaseReadersTestSuite struct {
	suite.Suite

	writer *io.PipeWriter
	conn   *MockedConnReader
}

func (suite *BaseReadersTestSuite) SetupTest() {
	reader, writer := io.Pipe()

	suite.writer = writer
	suite.conn = &MockedConnReader{reader: reader}
}

func (suite *BaseReadersTestSuite) TearDownTest() {
	suite.writer.Close()
	suite.conn.reader.Close()
}

type CountReaderTestSuite struct {
	BaseReadersTestSuite
}

func (suite *CountReaderTestSuite) TestReadByte() {
	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := countReader{
		conn:      suite.conn,
		bytesLeft: 1,
	}
	arr := make([]byte, 100)
	n, err := reader.Read(arr)

	suite.Equal(n, 1)
	suite.Equal(arr[:n], []byte{1})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.NotNil(err)
}

func (suite *CountReaderTestSuite) TestReadSeveralTimes() {
	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := countReader{
		conn:      suite.conn,
		bytesLeft: 5,
	}
	arr := make([]byte, 2)

	n, err := reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{1, 2})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{3, 4})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 1)
	suite.Equal(arr[:n], []byte{5})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.NotNil(err)
}

func (suite *CountReaderTestSuite) TestReadAll() {
	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := countReader{
		conn:      suite.conn,
		bytesLeft: 5,
	}
	arr := make([]byte, 10)

	n, err := reader.Read(arr)
	suite.Equal(n, 5)
	suite.Equal(arr[:n], []byte{1, 2, 3, 4, 5})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.NotNil(err)
}

type SimpleReaderTestSuite struct {
	BaseReadersTestSuite

	dialer *MockedDialer
}

func (suite *SimpleReaderTestSuite) SetupTest() {
	suite.BaseReadersTestSuite.SetupTest()
	suite.dialer = &MockedDialer{}
}

func (suite *SimpleReaderTestSuite) TestReadByte() {
	suite.dialer.On("Release", suite.conn, "addr")

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 1)
	arr := make([]byte, 10)

	n, err := reader.Read(arr)
	suite.Equal(n, 1)
	suite.Equal(arr[:n], []byte{1})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	suite.dialer.AssertExpectations(suite.T())
}

func (suite *SimpleReaderTestSuite) TestReadAll() {
	suite.dialer.On("Release", suite.conn, "addr")

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	arr := make([]byte, 10)

	n, err := reader.Read(arr)
	suite.Equal(n, 5)
	suite.Equal(arr[:n], []byte{1, 2, 3, 4, 5})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	suite.dialer.AssertExpectations(suite.T())
}

func (suite *SimpleReaderTestSuite) TestReadAllChunks() {
	suite.dialer.On("Release", suite.conn, "addr")

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	arr := make([]byte, 2)

	n, err := reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{1, 2})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{3, 4})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 1)
	suite.Equal(arr[:n], []byte{5})
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	suite.dialer.AssertExpectations(suite.T())
}

func (suite *SimpleReaderTestSuite) TestReadClose() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	arr := make([]byte, 2)

	n, err := reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{1, 2})
	suite.Nil(err)

	suite.writer.CloseWithError(errors.New("Unexpected"))

	n, err = reader.Read(arr)
	suite.NotNil(err)

	n, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	suite.dialer.AssertExpectations(suite.T())
	suite.conn.AssertExpectations(suite.T())
}

func (suite *SimpleReaderTestSuite) TestReadCloseEOF() {
	suite.dialer.On("Release", suite.conn, "addr")

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	arr := make([]byte, 2)

	n, err := reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{1, 2})
	suite.Nil(err)

	suite.writer.CloseWithError(io.EOF)

	n, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	n, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	suite.dialer.AssertExpectations(suite.T())
}

func (suite *SimpleReaderTestSuite) TestDoubleClose() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	suite.Nil(reader.Close())
	suite.Nil(reader.Close())

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *SimpleReaderTestSuite) TestCloseAfterError() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	arr := make([]byte, 2)
	n, err := reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{1, 2})
	suite.Nil(err)

	suite.writer.CloseWithError(errors.New("Unexpected"))
	n, err = reader.Read(arr)
	suite.Equal(n, 0)
	suite.NotNil(err)

	suite.Nil(reader.Close())

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *SimpleReaderTestSuite) TestCloseAfterEOF() {
	suite.dialer.On("Release", suite.conn, "addr")

	go func() {
		suite.writer.Write([]byte{1, 2, 3, 4, 5})
	}()

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5)
	arr := make([]byte, 2)
	n, err := reader.Read(arr)
	suite.Equal(n, 2)
	suite.Equal(arr[:n], []byte{1, 2})
	suite.Nil(err)

	suite.writer.CloseWithError(io.EOF)
	n, err = reader.Read(arr)
	suite.Equal(n, 0)
	suite.Equal(err, io.EOF)

	suite.Nil(reader.Close())

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "Release", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 0)
}

func TestCountReader(t *testing.T) {
	suite.Run(t, &CountReaderTestSuite{})
}

func TestSimpleReader(t *testing.T) {
	suite.Run(t, &SimpleReaderTestSuite{})
}
