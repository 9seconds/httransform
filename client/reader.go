package client

import (
	"context"
	"io"
)

type streamReader struct {
	data   io.Reader
	cancel context.CancelFunc
}

func (s *streamReader) Read(p []byte) (int, error) {
	n, err := s.data.Read(p)
	if err != nil {
		s.cancel()
	}

	return n, err
}

func (s *streamReader) Close() error {
	s.cancel()
	return nil
}
