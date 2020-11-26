package http

import (
	"io"
	"sync"
)

type closingReader struct {
	bufReader *bufferedReader
	reader    io.Reader
	closeOnce sync.Once
}

func (c *closingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	if err != nil {
		c.closeOnce.Do(c.doClose)
	}

	return n, err
}

func (c *closingReader) Close() error {
	c.closeOnce.Do(c.doClose)

	return nil
}

func (c *closingReader) doClose() {
    c.bufReader.callback()
    releaseBufferedReader(c.bufReader)
}
