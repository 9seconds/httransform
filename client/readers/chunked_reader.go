package readers

import (
	"bufio"
	"fmt"
	"io"

	"github.com/9seconds/httransform/v2/client/connectors"
	"github.com/PumpkinSeed/errors"
)

var chunkedReaderHex2IntTable = func() [256]byte {
	var b [256]byte

	for i := 0; i < 256; i++ {
		switch {
		case i >= '0' && i <= '9':
			b[i] = byte(i) - '0'
		case i >= 'a' && i <= 'f':
			b[i] = byte(i) - '0'
		case i >= 'a' && i <= 'f':
			b[i] = byte(i) - 'a' + 10
		case i >= 'A' && i <= 'F':
			b[i] = byte(i) - 'A' + 10
		default:
			b[i] = byte(16)
		}
	}

	return b
}()

type chunkedReader struct {
	baseReader

	readData bool
}

func (c *chunkedReader) Read(b []byte) (int, error) {
	if !c.readData {
		if err := c.prepareNextChunk(); err != nil {
			c.Close()
			return 0, err
		}
	}

	n, err := c.baseReader.Read(b)

	switch err {
	case nil:
		if c.baseReader.bytesLeft <= 0 {
			c.readData = false
			err = c.consumeCRLF()
			if err != nil {
				c.Close()
			}
		}
	case io.EOF:
		c.Release()
	default:
		c.Close()
	}

	if err != nil {
		err = errors.Wrap(err, ErrReader)
	}

	return n, err
}

func (c *chunkedReader) prepareNextChunk() error {
	size, err := c.readNextChunkSize()
	if err != nil {
		return err
	}

	if err := c.consumeCRLF(); err != nil {
		return err
	}

	if size == 0 {
		if c.consumeCRLF() != nil {
			c.Close()
		} else {
			c.Release()
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

		number := int64(chunkedReaderHex2IntTable[current])
		if number == 16 {
			if i == 0 {
				return -1, ErrReaderNoNumber
			}

			if err = c.baseReader.reader.UnreadByte(); err != nil {
				return -1, errors.Wrap(err, ErrReader)
			}

			return n, nil
		}

		if i > chunkedReaderMaxHexIntChars {
			return -1, ErrReaderTooLargeHexNum
		}

		n = (n << 4) | number
	}
}

func (c *chunkedReader) consumeCRLF() error {
	for {
		current, err := c.baseReader.reader.ReadByte()
		if err != nil {
			return errors.Wrap(err, ErrReaderCannotConsumeCRLF)
		}

		if current == '\r' {
			break
		}

		if current == ' ' {
			continue
		}

		return errors.Wrap(fmt.Errorf("expect '\r', got '%x'", current), ErrReaderUnexpectedCharacter)
	}

	current, err := c.baseReader.reader.ReadByte()
	if err != nil {
		return errors.Wrap(err, ErrReaderCannotConsumeCRLF)
	}

	if current != '\n' {
		return errors.Wrap(fmt.Errorf("expect '\n', got '%x'", current), ErrReaderUnexpectedCharacter)
	}

	return nil
}

func NewChunkedReader(conn connectors.Conn, reader *bufio.Reader, forceClose bool) Reader {
	return &chunkedReader{
		baseReader: baseReader{
			conn:       conn,
			reader:     reader,
			forceClose: forceClose,
		},
	}
}
