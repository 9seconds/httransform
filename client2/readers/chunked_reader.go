package readers

import (
	"fmt"
	"io"
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
		case i>= 'A' && i <='F':
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

		c.readData = true
	}

	n, err := c.baseReader.Read(b)

	switch err {
	case nil:
		if c.baseReader.bytesLeft <= 0 {
			c.readData = false
			if err := c.consumeCRLF(); err != nil {
				c.Close()
			}
		}
	case io.EOF:
		c.Release()
	default:
		c.Close()
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
				return -1, ErrReaderCannotUnread.WrapErr(err)
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
			return ErrReaderCannotConsumeCRLF.WrapErr(err)
		}

		if current == '\r' {
			break
		}

		if current == ' ' {
			continue
		}

		return ErrReaderUnexpectedCharacter.WrapErr(fmt.Errorf("expect '\r', got '%x'", current))
	}

	current, err := c.baseReader.reader.ReadByte()
	if err != nil {
		return ErrReaderCannotConsumeCRLF.WrapErr(err)
	}

	if current != '\n' {
		return ErrReaderUnexpectedCharacter.WrapErr(fmt.Errorf("expect '\n', got '%x'", current))
	}

	return nil
}
