package http

import (
	"bufio"
	"io"
	"sync"
)

type closingReader struct {
	bufReader *bufio.Reader
	reader    io.Reader
	closeOnce sync.Once
}

func (c *closingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	if err != nil {
		c.closeOnce.Do(c.doClose)
	}

	return n, err // nolint: wrapcheck
}

func (c *closingReader) Close() error {
	c.closeOnce.Do(c.doClose)

	return nil
}

func (c *closingReader) doClose() {
	releaseBufioReader(c.bufReader)
}
