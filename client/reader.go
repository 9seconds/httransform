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
	return s.data.Read(p)
}

func (s *streamReader) Close() error {
	s.cancel()
	return nil
}
