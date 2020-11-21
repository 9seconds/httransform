package headers

import (
	"bytes"
	"fmt"
	"net/textproto"
	"strings"
	"sync"
)

var newLineReplacer = strings.NewReplacer("\r\n", "\r\n\t")

type Header struct {
	ID    string
	Name  string
	Value string
}

type Headers struct {
	Headers []Header

	original FastHTTPHeaderWrapper
}

func (h *Headers) GetAll(key string) []*Header {
	key = makeHeaderID(key)
	rv := make([]*Header, 0, len(h.Headers))

	for i := range h.Headers {
		if key == h.Headers[i].ID {
			rv = append(rv, &h.Headers[i])
		}
	}

	return rv
}

func (h *Headers) GetAllExact(key string) []*Header {
	rv := make([]*Header, 0, len(h.Headers))

	for i := range h.Headers {
		if key == h.Headers[i].Name {
			rv = append(rv, &h.Headers[i])
		}
	}

	return rv
}

func (h *Headers) GetLast(key string) *Header {
	headers := h.GetAll(key)

	if len(headers) == 0 {
		return headers[len(headers)-1]
	}

	return nil
}

func (h *Headers) GetLastExact(key string) *Header {
	headers := h.GetAllExact(key)

	if len(headers) == 0 {
		return headers[len(headers)-1]
	}

	return nil
}

func (h *Headers) Set(name, value string, cleanupRest bool) {
	found := false
	key := makeHeaderID(name)

	for i := 0; i < len(h.Headers); i++ {
		if key == h.Headers[i].ID {
			if cleanupRest && found {
				copy(h.Headers[i:], h.Headers[i+1:])
				h.Headers = h.Headers[:len(h.Headers)-1]
			} else if !found {
				h.Headers[i].Name = name
				h.Headers[i].Value = value
				found = true
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
		if name == h.Headers[i].Name {
			if cleanupRest && found {
				copy(h.Headers[i:], h.Headers[i+1:])
				h.Headers = h.Headers[:len(h.Headers)-1]
			} else if !found {
				h.Headers[i].Value = value
				found = true
			}
		}
	}

	if !found {
		h.Append(name, value)
	}
}

func (h *Headers) Remove(key string) {
	key = makeHeaderID(key)

	for i := 0; i < len(h.Headers); i++ {
		if key == h.Headers[i].ID {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		}
	}
}

func (h *Headers) RemoveExact(key string) {
	for i := 0; i < len(h.Headers); i++ {
		if key == h.Headers[i].Name {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		}
	}
}

func (h *Headers) Append(name, value string) {
	h.Headers = append(h.Headers, Header{
		ID:    makeHeaderID(name),
		Name:  name,
		Value: value,
	})
}

func (h *Headers) Sync() error {
	if h.original == nil {
		return nil
	}

	h.original.DisableNormalizing()

	buf := bytes.Buffer{}

	if err := h.syncWriteFirstLine(&buf); err != nil {
		return fmt.Errorf("cannot get a first line: %w", err)
	}

	h.syncWriteHeaders(&buf)

	if err := h.original.Read(&buf); err != nil {
		return fmt.Errorf("cannot parse given headers: %w", err)
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
		buf.WriteString(h.Headers[i].Name)
		buf.WriteByte(':')
		buf.WriteByte(' ')
		newLineReplacer.WriteString(buf, h.Headers[i].Value) // nolint: errcheck
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