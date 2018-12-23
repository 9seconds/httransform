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

	if err == errCountReaderExhaust {
		err = io.EOF
	}
	if err == io.EOF {
		s.dialer.Release(s.countReader.conn.(net.Conn), s.addr)
		s.closed = true
	} else {
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

	if !c.readData {
		if err := c.prepareNextChunk(); err != nil {
			c.Close()
			return 0, err
		}
		c.readData = true
	}

	n, err := c.countReader.Read(b)
	if c.countReader.bytesLeft <= 0 {
		c.readData = false
		err = c.consumeCRLF()
		if err != nil {
			c.Close()
		}
	}
	return n, err
}

func (c *chunkedReader) Close() error {
	if !c.closed {
		poolBufferedReader.Put(c.bufferedReader)
		c.conn.Close()
		c.dialer.NotifyClosed(c.addr)
		c.closed = true
	}

	return nil
}

func (c *chunkedReader) prepareNextChunk() error {
	size, err := c.readNextChunkSize()
	if err != nil {
		return err
	}

	if err = c.consumeCRLF(); err != nil {
		return err
	}

	if size == 0 {
		if c.consumeCRLF() != nil {
			c.Close()
		} else {
			c.dialer.Release(c.conn, c.addr)
			c.closed = true
			poolBufferedReader.Put(c.bufferedReader)
		}
		return io.EOF
	}
	c.countReader.bytesLeft = size

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
		if current == ' ' {
			continue
		}
		if current == '\r' {
			break
		}
		return errors.Errorf("Expect '\r', got '%x'", current)
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
	bufferedReader := poolBufferedReader.Get().(*bufio.Reader)
	bufferedReader.Reset(conn)

	return &chunkedReader{
		countReader:    countReader{conn: bufferedReader},
		dialer:         dialer,
		conn:           conn,
		addr:           addr,
		bufferedReader: bufferedReader,
	}
}
