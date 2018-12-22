package client

import (
	"bufio"
	"io"
	"net"

	"github.com/juju/errors"
)

var (
	errCountReaderExhaust = errors.New("Cannot read more")
	errTooLargeHexNum     = errors.New("Too large hex number")
	errNoNumber           = errors.New("Cannot find a number")

	hex2intTable = func() [256]byte {
		var b [256]byte

		for i := 0; i < 256; i++ {
			c := byte(16)
			if i >= '0' && i <= '9' {
				c = byte(i) - '0'
			} else if i >= 'a' && i <= 'f' {
				c = byte(i) - 'a' + 10
			} else if i >= 'A' && i <= 'F' {
				c = byte(i) - 'A' + 10
			}
			b[i] = c
		}

		return b
	}()
)

const (
	// https://golang.org/ref/spec#Numeric_types
	maxInt64Value           = 9223372036854775807
	chunkedReaderBufferSize = 16 * 1024
)

type countReader struct {
	conn      io.Reader
	bytesLeft int64
}

func (c *countReader) Read(b []byte) (int, error) {
	if c.bytesLeft <= 0 {
		return 0, errCountReaderExhaust
	}

	if int64(len(b)) > c.bytesLeft {
		b = b[:c.bytesLeft]
	}

	n, err := c.conn.Read(b)
	c.bytesLeft -= int64(n)

	return n, err
}

type simpleReader struct {
	countReader

	dialer Dialer
	addr   string
	closed bool
}

func (s *simpleReader) Read(b []byte) (int, error) {
	if s.closed {
		return 0, io.EOF
	}

	n, err := s.countReader.Read(b)
	if err == nil {
		return n, err
	}

	switch err {
	case errCountReaderExhaust, io.EOF:
		s.dialer.Release(s.countReader.conn.(net.Conn), s.addr)
		s.closed = true
	default:
		s.Close()
	}

	return n, err
}

func (s *simpleReader) Close() error {
	if !s.closed {
		s.countReader.conn.(io.Closer).Close()
		s.dialer.NotifyClosed(s.addr)
		s.closed = true
	}

	return nil
}

type chunkedReader struct {
	countReader

	dialer         Dialer
	conn           net.Conn
	addr           string
	readData       bool
	closed         bool
	bufferedReader *bufio.Reader
}

func (c *chunkedReader) Read(b []byte) (int, error) {
	if c.closed {
		return 0, io.EOF
	}

	if c.readData {
		n, err := c.countReader.Read(b)
		if err == errCountReaderExhaust {
			c.readData = false
			err = c.consumeCRLF()
		}
		return n, err
	}

	size, err := c.readNextChunkSize()
	if err != nil {
		return 0, err
	}

	err = c.consumeCRLF()
	if err != nil {
		return 0, err
	}

	if size == 0 {
		c.closed = true
		if c.consumeCRLF() != nil {
			c.conn.Close()
			c.dialer.NotifyClosed(c.addr)
		} else {
			c.dialer.Release(c.conn, c.addr)
		}
		return 0, io.EOF
	}

	c.countReader.bytesLeft = size
	c.readData = true
}

func (c *chunkedReader) Close() error {
	if !c.closed {
		c.conn.Close()
		c.dialer.NotifyClosed(c.addr)
		c.closed = true
	}

	return nil
}

func (c *chunkedReader) readNextChunkSize() (int64, error) {
	var n int64
	for i := 0; ; i++ {
		current, err := c.bufferedReader.ReadByte()
		if err != nil {
			return -1, err
		}

		number := int64(hex2intTable[current])
		if number == 16 {
			if i == 0 {
				return -1, errNoNumber
			}
			c.bufferedReader.UnreadByte()
			return n, nil
		}
		if i > maxHexIntChars {
			return -1, errTooLargeHexNum
		}

		n = (n << 4) | number
	}
}

func (c *chunkedReader) consumeCRLF() error {
	for {
		current, err := c.bufferedReader.ReadByte()
		if err != nil {
			return errors.Annotate(err, "Cannot consume CRLF")
		}

		switch current {
		case ' ':
			continue
		case '\r':
			break
		default:
			return errors.Errorf("Expect '\r', got '%x'", current)
		}
	}

	current, err := c.bufferedReader.ReadByte()
	if err != nil {
		return errors.Annotate(err, "Cannot consume CRLF")
	}
	if current != '\n' {
		return errors.Errorf("Expect '\n', got '%x'", current)
	}

	return nil
}

func newSimpleReader(addr string, conn net.Conn, dialer Dialer, contentLength int64) *simpleReader {
	return &simpleReader{
		countReader: countReader{
			conn:      conn,
			bytesLeft: contentLength,
		},
		dialer: dialer,
		addr:   addr,
	}
}

func newChunkedReader(addr string, conn net.Conn, dialer Dialer) *chunkedReader {
	bufferedReader := bufio.NewReaderSize(conn, chunkedReaderBufferSize)

	return &chunkedReader{
		countReader:    countReader{conn: bufferedReader},
		dialer:         dialer,
		conn:           conn,
		addr:           addr,
		bufferedReader: bufferedReader,
	}
}
