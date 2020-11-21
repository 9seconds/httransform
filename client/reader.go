package client

import (
	"io"
	"sync"
)

type streamReader struct {
	bufferedConn *bufferedConn
	toReadFrom   io.Reader
	closeOnce    sync.Once
}

func (s *streamReader) Read(p []byte) (int, error) {
	n, err := s.toReadFrom.Read(p)
	if err != nil {
		s.Close()
	}

	return n, err // nolint: wrapcheck
}

func (s *streamReader) Close() error {
	s.closeOnce.Do(func() {
		s.bufferedConn.Cancel()
		releaseBufferedConn(s.bufferedConn)
	})

	return nil
}
