package client

import (
	"bufio"
	"io"
	"net"

	"golang.org/x/xerrors"
)

var (
	errCountReaderExhaust = xerrors.New("cannot read more")
	errTooLargeHexNum     = xerrors.New("too large hex number")
	errNoNumber           = xerrors.New("cannot find a number")

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

type baseReader struct {
	reader      *bufio.Reader
	conn        net.Conn
	dialer      Dialer
	addr        string
	bytesLeft   int64
	closed      bool
	alwaysClose bool
}

func (b *baseReader) Read(p []byte) (int, error) {
	if b.bytesLeft <= 0 {
		return 0, errCountReaderExhaust
	}

	if int64(len(p)) > b.bytesLeft {
		p = p[:b.bytesLeft]
	}

	n, err := b.reader.Read(p)
	b.bytesLeft -= int64(n)

	return n, err
}

func (b *baseReader) releaseConn() {
	if b.alwaysClose {
		b.closeConn()
	} else if !b.closed {
		b.dialer.Release(b.conn, b.addr)
		b.closed = true
		b.reader.Reset(nil)
		poolBufferedReader.Put(b.reader)
	}
}

func (b *baseReader) closeConn() {
	if !b.closed {
		b.conn.Close()
		b.dialer.NotifyClosed(b.addr)
		b.closed = true
		b.reader.Reset(nil)
		poolBufferedReader.Put(b.reader)
	}
}

type simpleReader struct {
	baseReader
}

func (s *simpleReader) Read(b []byte) (int, error) {
	if s.closed {
		return 0, io.EOF
	}

	n, err := s.baseReader.Read(b)
	if err == nil {
		return n, err
	}

	if err == errCountReaderExhaust {
		err = io.EOF
	}

	if err == io.EOF {
		s.releaseConn()
	} else {
		s.closeConn()
	}

	return n, err
}

func (s *simpleReader) Close() error {
	s.closeConn()
	return nil
}

type chunkedReader struct {
	baseReader

	readData bool
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

	n, err := c.baseReader.Read(b)

	if c.baseReader.bytesLeft <= 0 {
		c.readData = false
		err = c.consumeCRLF()

		if err != nil {
			c.Close()
		}
	}

	return n, err
}

func (c *chunkedReader) Close() error {
	c.closeConn()
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
			c.releaseConn()
		}

		return io.EOF
	}

	c.baseReader.bytesLeft = size

	return nil
}

func (c *chunkedReader) readNextChunkSize() (int64, error) {
	var n int64

	for i := 0; ; i++ {
		current, err := c.baseReader.reader.ReadByte()
		if err != nil {
			return -1, err
		}

		number := int64(hex2intTable[current])
		if number == 16 {
			if i == 0 {
				return -1, errNoNumber
			}

			if err = c.baseReader.reader.UnreadByte(); err != nil {
				return -1, xerrors.Errorf("cannot unread byte: %w", err)
			}

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
		current, err := c.baseReader.reader.ReadByte()
		if err != nil {
			return xerrors.Errorf("cannot consume CRLF: %w", err)
		}

		if current == ' ' {
			continue
		}

		if current == '\r' {
			break
		}

		return xerrors.Errorf("Expect '\r', got '%x'", current)
	}

	current, err := c.baseReader.reader.ReadByte()
	if err != nil {
		return xerrors.Errorf("cannot consume CRLF: %w", err)
	}

	if current != '\n' {
		return xerrors.Errorf("Expect '\n', got '%x'", current)
	}

	return nil
}

func newSimpleReader(addr string, conn net.Conn, reader *bufio.Reader, dialer Dialer, closeOnEOF bool, contentLength int64) *simpleReader {
	return &simpleReader{
		baseReader: baseReader{
			reader:      reader,
			conn:        conn,
			dialer:      dialer,
			addr:        addr,
			bytesLeft:   contentLength,
			alwaysClose: closeOnEOF,
		},
	}
}

func newChunkedReader(addr string, conn net.Conn, reader *bufio.Reader, dialer Dialer, closeOnEOF bool) *chunkedReader {
	return &chunkedReader{
		baseReader: baseReader{
			reader:      reader,
			conn:        conn,
			dialer:      dialer,
			addr:        addr,
			alwaysClose: closeOnEOF,
		},
	}
}
