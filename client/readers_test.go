package client

import (
	"bytes"
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 1, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

	reader := newSimpleReader("addr", suite.conn, suite.dialer, 5, false)
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

type ChunkedReaderTestSuite struct {
	BaseReadersTestSuite

	dialer *MockedDialer
}

func (suite *ChunkedReaderTestSuite) SetupTest() {
	suite.BaseReadersTestSuite.SetupTest()
	suite.dialer = &MockedDialer{}
}

func (suite *ChunkedReaderTestSuite) TestReferenceExample() {
	suite.dialer.On("Release", suite.conn, "addr")

	data := "4\r\nWiki\r\n5\r\npedia\r\nE\r\n in\r\n\r\nchunks.\r\n0\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 4)
	suite.Equal(arr[:n], []byte("Wiki"))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 5)
	suite.Equal(arr[:n], []byte("pedia"))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, len([]byte(" in\r\n\r\nchunks.")))
	suite.Equal(arr[:n], []byte(" in\r\n\r\nchunks."))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 0)
	suite.Equal(err, io.EOF)
}

func (suite *ChunkedReaderTestSuite) TestMDNExample() {
	suite.dialer.On("Release", suite.conn, "addr")

	data := "7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 7)
	suite.Equal(arr[:n], []byte("Mozilla"))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 9)
	suite.Equal(arr[:n], []byte("Developer"))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 7)
	suite.Equal(arr[:n], []byte("Network"))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 0)
	suite.Equal(err, io.EOF)
}

func (suite *ChunkedReaderTestSuite) TestBigNumber() {
	suite.dialer.On("Release", suite.conn, "addr")

	data := "1a\r\n" + string(bytes.Repeat([]byte{'a'}, 26)) + "\r\n0\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 26)
	suite.Equal(arr[:n], bytes.Repeat([]byte{'a'}, 26))
	suite.Nil(err)

	n, err = reader.Read(arr)
	suite.Equal(n, 0)
	suite.Equal(err, io.EOF)
}

func (suite *ChunkedReaderTestSuite) TestTooBigNumber() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	data := "1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\r\n0\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 0)
	suite.NotNil(err)

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *ChunkedReaderTestSuite) TestNoNumber() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	data := "\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 0)
	suite.NotNil(err)

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *ChunkedReaderTestSuite) TestCorruptedNumber() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	data := "5\r\t\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 0)
	suite.NotNil(err)

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *ChunkedReaderTestSuite) TestCorruptedNumber2() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	data := "5\n\t\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 0)
	suite.NotNil(err)

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *ChunkedReaderTestSuite) TestCorruptedCRLF() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	data := "5\r\nhello\n\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	n, err := reader.Read(arr)
	suite.Equal(n, 5)
	suite.NotNil(err)

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *ChunkedReaderTestSuite) TestSlowReader() {
	suite.dialer.On("Release", suite.conn, "addr")

	data := "7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()
	consumed := []byte{}

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 2)
	var err error
	var n int
	for err == nil {
		n, err = reader.Read(arr)
		consumed = append(consumed, arr[:n]...)
	}
	suite.Equal(consumed, []byte("MozillaDeveloperNetwork"))
	suite.Equal(err, io.EOF)
}

func (suite *ChunkedReaderTestSuite) TestCloseUnexpected() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	data := "7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\n\r\n"
	go func() {
		suite.writer.Write([]byte(data))
	}()

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	arr := make([]byte, 100)
	reader.Read(arr)

	suite.writer.CloseWithError(errors.New("Unexpected"))
	reader.bufferedReader.Reset(suite.conn)

	_, err := reader.Read(arr)
	suite.NotNil(err)

	_, err = reader.Read(arr)
	suite.Equal(err, io.EOF)

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func (suite *ChunkedReaderTestSuite) TestDoubleClose() {
	suite.dialer.On("NotifyClosed", "addr")
	suite.conn.On("Close").Return(nil)

	reader := newChunkedReader("addr", suite.conn, suite.dialer, false)
	suite.Nil(reader.Close())
	suite.Nil(reader.Close())

	suite.dialer.AssertExpectations(suite.T())
	suite.dialer.AssertNumberOfCalls(suite.T(), "NotifyClosed", 1)
	suite.conn.AssertExpectations(suite.T())
	suite.conn.AssertNumberOfCalls(suite.T(), "Close", 1)
}

func TestCountReader(t *testing.T) {
	suite.Run(t, &CountReaderTestSuite{})
}

func TestSimpleReader(t *testing.T) {
	suite.Run(t, &SimpleReaderTestSuite{})
}

func TestChunkedReader(t *testing.T) {
	suite.Run(t, &ChunkedReaderTestSuite{})
}
