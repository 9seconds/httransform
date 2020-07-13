package readers

import (
	"bufio"
	"io"

	"github.com/PumpkinSeed/errors"

	"github.com/9seconds/httransform/v2/client/connectors"
)

type simpleReader struct {
	baseReader
}

func (s *simpleReader) Read(p []byte) (int, error) {
	n, err := s.baseReader.Read(p)

	switch err {
	case nil:
		return n, nil
	case io.EOF:
		s.Release()
	default:
		s.Close()
	}

	return n, errors.Wrap(err, ErrReader)
}

func NewSimpleReader(conn connectors.Conn, reader *bufio.Reader, forceClose bool, length int64) Reader {
	return &simpleReader{
		baseReader: baseReader{
			conn:       conn,
			reader:     reader,
			forceClose: forceClose,
			bytesLeft:  length,
		},
	}
}
