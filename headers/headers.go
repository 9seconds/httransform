package headers

import (
	"bytes"
	"net/textproto"
	"strings"
	"sync"

	"github.com/9seconds/httransform/v2/errors"
)

var newLineReplacer = strings.NewReplacer("\r\n", "\r\n\t")

// Headers present a list of Header with some convenient optimized
// shorthand operations. It also has a binding to corresponding fasthttp
// Header interface for synchronisation.
//
// Once you instantiate a new Headers instance with either
// AcquireHeaderSet or simple headers.Headers{}, you need to fill it
// with values. Use Init function to do that. Each operation is not
// reflected to underlying fasthttp Header. To do that, you have to
// implicitly call Sync methods.
type Headers struct {
	// A list of headers. Actually, you do not need to use all
	// operations from Headers struct to manage this list. These are
	// shortcuts. If you like to manipulate with array, nothing really
	// prevents you from going this way.
	Headers []Header

	original FastHTTPHeaderWrapper
}

// Strings here to comply with fmt.Stringer interface.
func (h *Headers) String() string {
	if len(h.Headers) == 0 {
		return ""
	}

	builder := strings.Builder{}

	h.Headers[0].writeString(&builder)

	for i := 1; i < len(h.Headers); i++ {
		builder.WriteRune('\n')
		h.Headers[i].writeString(&builder)
	}

	return builder.String()
}

// GetAll returns a list of headers where name is given in a
// case-INSENSITIVE fashion.
func (h *Headers) GetAll(key string) []*Header {
	keyID := makeHeaderID(key)
	rv := make([]*Header, 0, len(h.Headers))

	for i := range h.Headers {
		if h.Headers[i].id == keyID {
			rv = append(rv, &h.Headers[i])
		}
	}

	return rv
}

// GetAllExact returns a list of headers where name is given in a
// case-SENSITIVE fashion.
func (h *Headers) GetAllExact(key string) []*Header {
	rv := make([]*Header, 0, len(h.Headers))

	for i := range h.Headers {
		if key == h.Headers[i].name {
			rv = append(rv, &h.Headers[i])
		}
	}

	return rv
}

// GetLast returns a last header (or nil) where name is
// case-INSENSITIVE.
func (h *Headers) GetLast(key string) *Header {
	keyID := makeHeaderID(key)

	for i := len(h.Headers) - 1; i >= 0; i-- {
		if h.Headers[i].id == keyID {
			return &h.Headers[i]
		}
	}

	return nil
}

// GetLastExact returns a last header (or nil) where name is
// case-SENSITIVE.
func (h *Headers) GetLastExact(key string) *Header {
	for i := len(h.Headers) - 1; i >= 0; i-- {
		if key == h.Headers[i].name {
			return &h.Headers[i]
		}
	}

	return nil
}

// GetLast returns a first header (or nil) where name is
// case-INSENSITIVE.
func (h *Headers) GetFirst(key string) *Header {
	keyID := makeHeaderID(key)

	for i := range h.Headers {
		if h.Headers[i].id == keyID {
			return &h.Headers[i]
		}
	}

	return nil
}

// GetLast returns a first header (or nil) where name is
// case-SENSITIVE.
func (h *Headers) GetFirstExact(key string) *Header {
	for i := range h.Headers {
		if key == h.Headers[i].name {
			return &h.Headers[i]
		}
	}

	return nil
}

// Set sets a value for the FIRST seen header where name is
// case-INSENSITIVE. cleanupRest parameter defines if the rest values
// should be wiped out.
func (h *Headers) Set(name, value string, cleanupRest bool) {
	found := false
	key := makeHeaderID(name)

	for i := 0; i < len(h.Headers); i++ {
		if h.Headers[i].id != key {
			continue
		}

		switch {
		case cleanupRest && found:
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		case cleanupRest && !found:
			h.Headers[i].value = value
			found = true
		case !found:
			h.Headers[i].value = value

			return
		}
	}

	if !found {
		h.Append(name, value)
	}
}

// SetExact sets a value for the FIRST seen header where name is
// case-SENSITIVE. cleanupRest parameter defines if the rest values
// should be wiped out.
func (h *Headers) SetExact(name, value string, cleanupRest bool) {
	found := false

	for i := 0; i < len(h.Headers); i++ {
		if name != h.Headers[i].name {
			continue
		}

		switch {
		case cleanupRest && found:
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		case cleanupRest && !found:
			h.Headers[i].value = value
			found = true
		case !found:
			h.Headers[i].value = value

			return
		}
	}

	if !found {
		h.Append(name, value)
	}
}

// Remove removes all headers from the list where name is
// case-INSENSITIVE. Order is kept.
func (h *Headers) Remove(key string) {
	keyID := makeHeaderID(key)

	for i := 0; i < len(h.Headers); {
		if h.Headers[i].id == keyID {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		} else {
			i++
		}
	}
}

// Remove removes all headers from the list where name is
// case-SENSITIVE. Order is kept.
func (h *Headers) RemoveExact(key string) {
	for i := 0; i < len(h.Headers); {
		if key == h.Headers[i].name {
			copy(h.Headers[i:], h.Headers[i+1:])
			h.Headers = h.Headers[:len(h.Headers)-1]
		} else {
			i++
		}
	}
}

// Append simply appends a header to the list of headers.
func (h *Headers) Append(name, value string) {
	h.Headers = append(h.Headers, Header{
		id:    makeHeaderID(name),
		name:  name,
		value: value,
	})
}

// Reset rebinds Headers to corresponding fasthttp data structure.
func (h *Headers) Reset(original FastHTTPHeaderWrapper) {
	h.Headers = h.Headers[:0]
	h.original = original
}

// Push resets underlying fasthttp Header structure and restores it
// based on a given header list.
//
// Ananlogy: git push where origin is fasthttp.Header.
func (h *Headers) Push() error {
	if h.original == nil {
		return nil
	}

	h.original.DisableNormalizing()

	buf := acquireBytesBuffer()
	defer releaseBytesBuffer(buf)

	if err := h.pushFirstLine(buf); err != nil {
		return errors.Annotate(err, "cannot get a first line", "headers_sync", 0)
	}

	h.pushHeaders(buf)

	if err := h.original.Read(buf); err != nil {
		return errors.Annotate(err, "cannot parse given headers", "headers_sync", 0)
	}

	for _, v := range h.GetLast("Connection").Values() {
		if v != "close" {
			h.original.ResetConnectionClose()
		}
	}

	return nil
}

func (h *Headers) pushFirstLine(buf *bytes.Buffer) error {
	bytesReader := acquireBytesReader(h.original.Headers())
	defer releaseBytesReader(bytesReader)

	bufReader := acquireBufioReader(bytesReader)
	defer releaseBufioReader(bufReader)

	reader := textproto.Reader{R: bufReader}

	line, err := reader.ReadLineBytes()
	if err != nil {
		return errors.Annotate(err, "cannot scan a first line", "headers_sync", 0)
	}

	buf.Write(line)
	buf.WriteByte('\r')
	buf.WriteByte('\n')

	return nil
}

func (h *Headers) pushHeaders(buf *bytes.Buffer) {
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

// Pull reads underlying fasthttp.Header and fills a header list with
// its contents.
//
// Ananlogy: git pull where origin is fasthttp.Header.
func (h *Headers) Pull() error { // nolint: cyclop
	h.Headers = h.Headers[:0]

	if h.original == nil {
		return nil
	}

	bytesReader := acquireBytesReader(h.original.Headers())
	defer releaseBytesReader(bytesReader)

	bufReader := acquireBufioReader(bytesReader)
	defer releaseBufioReader(bufReader)

	reader := textproto.Reader{
		R: bufReader,
	}

	if _, err := reader.ReadLineBytes(); err != nil {
		return errors.Annotate(err, "cannot skip a first line", "headers_init", 0)
	}

	for {
		line, err := reader.ReadContinuedLineBytes()
		if len(line) == 0 {
			return nil
		}

		if err != nil {
			return errors.Annotate(err, "cannot parse headers", "headers_init", 0)
		}

		colPosition := bytes.IndexByte(line, ':')
		if colPosition < 0 {
			return errors.Annotate(err, "malformed header "+string(line), "headers_init", 0)
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

// AcquireHeaderSet takes a new Header instance from the pool. Please do
// not remember to return it back when you are done and do not use any
// references on that after.
func AcquireHeaderSet() *Headers {
	return poolHeaderSet.Get().(*Headers)
}

// ReleaseHeaderSet resets a Headers and returns it back to a pool.
// You should not use an instance of that header after this operation.
func ReleaseHeaderSet(set *Headers) {
	set.Reset(nil)
	poolHeaderSet.Put(set)
}
