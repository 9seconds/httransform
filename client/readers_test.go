package client

import (
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CountReaderTestSuite struct {
	suite.Suite

	writer *io.PipeWriter
	conn   *MockedConnReader
}

func (suite *CountReaderTestSuite) SetupTest() {
	reader, writer := io.Pipe()

	suite.writer = writer
	suite.conn = &MockedConnReader{reader: reader}
}

func (suite *CountReaderTestSuite) TearDownTest() {
	suite.writer.Close()
	suite.conn.reader.Close()
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

func TestCountReader(t *testing.T) {
	suite.Run(t, &CountReaderTestSuite{})
}
