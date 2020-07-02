package readers

import "io"

type simpleReader struct {
	baseReader
}

func (s *simpleReader) Read(b []byte) (int, error) {
	n, err := s.baseReader.Read(b)

	switch err {
	case nil:
	case io.EOF:
		s.Release()
	default:
		s.Close()
	}

	return n, err
}
