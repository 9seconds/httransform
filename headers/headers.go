package headers

import (
	"bytes"
	"fmt"
	"net/textproto"
	"strings"
	"sync"
)

var newLineReplacer = strings.NewReplacer("\r\n", "\r\n\t")

type Headers struct {
	Headers []Header

	original FastHTTPHeaderWrapper
}

func (h *Headers) GetAll(key string) []*Header {
	bkey := makeHeaderID(key)
	rv := make([]*Header, 0, len(h.Headers))

	for i := range h.Headers {
		if bytes.Equal(bkey, h.Headers[i].ID()) {
			rv = append(rv, &h.Headers[i])
		}
	}

	return rv
}

func (h *Headers) GetAllExact(key string) []*Header {
	rv := make([]*Header, 0, len(h.Headers))

	for i := range h.Headers {
		if key == h.Headers[i].name {
			rv = append(rv, &h.Headers[i])
		}
	}

	return rv
}

func (h *Headers) GetLast(key string) *Header {
	bkey := makeHeaderID(key)

	for i := len(h.Headers) - 1; i >= 0; i-- {
		if bytes.Equal(bkey, h.Headers[i].ID()) {
			return &h.Headers[i]
		}
	}

	return nil
}

func (h *Headers) GetLastExact(key string) *Header {
	for i := len(h.Headers) - 1; i >= 0; i-- {
		if key == h.Headers[i].name {
			return &h.Headers[i]
		}
	}

	return nil
}

func (h *Headers) GetFirst(key string) *Header {
	bkey := makeHeaderID(key)

	for i := range h.Headers {
		if bytes.Equal(bkey, h.Headers[i].ID()) {
			return &h.Headers[i]
		}
	}

	return nil
}

func (h *Headers) GetFirstExact(key string) *Header {
	for i := range h.Headers {
		if key == h.Headers[i].name {
			return &h.Headers[i]
		}
	}

	return nil
}

func (h *Headers) Set(name, value string, cleanupRest bool) {
	found := false
	key := makeHeaderID(name)

	for i := 0; i < len(h.Headers); i++ {
		if !bytes.Equal(key, h.Headers[i].ID()) {
			continue
		}

		if cleanupRest && found {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		} else if !found {
			found = true
			h.Headers[i].value = value

			if name != h.Headers[i].name {
				h.Headers[i].name = name
				h.Headers[i].id = makeHeaderID(name)
			}
		}
	}

	if !found {
		h.Append(name, value)
	}
}

func (h *Headers) SetExact(name, value string, cleanupRest bool) {
	found := false

	for i := 0; i < len(h.Headers); i++ {
		if name != h.Headers[i].name {
			continue
		}

		if cleanupRest && found {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		} else if !found {
			h.Headers[i].value = value
			found = true
		}
	}

	if !found {
		h.Append(name, value)
	}
}

func (h *Headers) Remove(key string) {
    bkey := makeHeaderID(key)

	for i := 0; i < len(h.Headers); i++ {
		if bytes.Equal(bkey, h.Headers[i].ID()) {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		}
	}
}

func (h *Headers) RemoveExact(key string) {
	for i := 0; i < len(h.Headers); i++ {
		if key == h.Headers[i].name {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		}
	}
}

func (h *Headers) Append(name, value string) {
	h.Headers = append(h.Headers, NewHeader(name, value))
}

func (h *Headers) Sync() error {
	if h.original == nil {
		return nil
	}

	h.original.DisableNormalizing()

	buf := acquireBytesBuffer()
	defer releaseBytesBuffer(buf)

	if err := h.syncWriteFirstLine(buf); err != nil {
		return fmt.Errorf("cannot get a first line: %w", err)
	}

	h.syncWriteHeaders(buf)

	if err := h.original.Read(buf); err != nil {
		return fmt.Errorf("cannot parse given headers: %w", err)
	}

	for _, v := range h.GetLast("Connection").Values() {
		if v != "close" {
			h.original.ResetConnectionClose()
		}
	}

	return nil
}

func (h *Headers) syncWriteFirstLine(buf *bytes.Buffer) error {
	bytesReader := acquireBytesReader(h.original.Headers())
	defer releaseBytesReader(bytesReader)

	bufReader := acquireBufioReader(bytesReader)
	defer releaseBufioReader(bufReader)

	reader := textproto.Reader{R: bufReader}

	line, err := reader.ReadLineBytes()
	if err != nil {
		return fmt.Errorf("cannot scan a first line: %w", err)
	}

	buf.Write(line)
	buf.WriteByte('\r')
	buf.WriteByte('\n')

	return nil
}

func (h *Headers) syncWriteHeaders(buf *bytes.Buffer) {
	for i := range h.Headers {
		buf.WriteString(h.Headers[i].name)
		buf.WriteByte(':')
		buf.WriteByte(' ')
		newLineReplacer.WriteString(buf, h.Headers[i].value) // nolint: errcheck
		buf.WriteByte('\r')
		buf.WriteByte('\n')
	}

	buf.WriteByte('\r')
	buf.WriteByte('\n')
}

func (h *Headers) Reset() {
	h.Headers = h.Headers[:0]
	h.original = nil
}

func (h *Headers) Init(original FastHTTPHeaderWrapper) error {
	h.Reset()

	if original == nil {
		return nil
	}

	h.original = original

	bytesReader := acquireBytesReader(original.Headers())
	defer releaseBytesReader(bytesReader)

	bufReader := acquireBufioReader(bytesReader)
	defer releaseBufioReader(bufReader)

	reader := textproto.Reader{
		R: bufReader,
	}

	if _, err := reader.ReadLineBytes(); err != nil {
		return fmt.Errorf("cannot skip first line: %w", err)
	}

	for {
		line, err := reader.ReadContinuedLineBytes()
		if len(line) == 0 {
			return nil
		}

		if err != nil {
			h.Reset()

			return fmt.Errorf("cannot parse headers: %w", err)
		}

		colPosition := bytes.IndexByte(line, ':')
		if colPosition < 0 {
			h.Reset()

			return fmt.Errorf("malformed header %s", string(line))
		}

		endKey := colPosition
		for endKey > 0 && line[endKey-1] == ' ' {
			endKey--
		}

		colPosition++

		for colPosition < len(line) && (line[colPosition] == ' ' || line[colPosition] == '\t') {
			colPosition++
		}

		h.Append(string(line[:endKey]), string(line[colPosition:]))
	}
}

var poolHeaderSet = sync.Pool{
	New: func() interface{} {
		return &Headers{
			Headers: []Header{},
		}
	},
}

func AcquireHeaderSet() *Headers {
	return poolHeaderSet.Get().(*Headers)
}

func ReleaseHeaderSet(set *Headers) {
	set.Reset()
	poolHeaderSet.Put(set)
}
