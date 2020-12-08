package headers

import (
	"io"
	"net/textproto"
	"strings"
)

// Header presents a single header instance with a name or value.
// Nothing is explicitly exposed because we do some additional actions
// to speedup some operations.
type Header struct {
	id    string
	value string
	name  string
}

// String complies with fmt.Stringer interface.
func (h *Header) String() string {
	builder := strings.Builder{}

	h.writeString(&builder)

	return builder.String()
}

// ID returns a unique identifier of the header.
func (h *Header) ID() string {
	if h == nil {
		return ""
	}

	return h.id
}

// Name returns a header name. Case-sensitive.
func (h *Header) Name() string {
	if h == nil {
		return ""
	}

	return h.name
}

// CanonicalName name returns a name in canonical form: dash-separated
// parts are titlecased. For example, a canonical form for
// 'aCCept-Encoding' is 'Accept-Encoding'.
func (h *Header) CanonicalName() string {
	return textproto.CanonicalMIMEHeaderKey(h.Name())
}

// Value returns a value of the header.
func (h *Header) Value() string {
	if h == nil {
		return ""
	}

	return h.value
}

// Values returns a list of header values. If name is a comma-delimited
// list, it will return a list of parts. For example, for
// header 'Connection: keep-alive, Upgrade', this one returns
// []string{"keep-alive", "Upgrade"}.
//
// This is a shortcut to headers.Values(h.Value()).
func (h *Header) Values() []string {
	return Values(h.Value())
}

// WithName returns a copy of header with modified name.
func (h *Header) WithName(name string) Header {
	var value string

	if h != nil {
		value = h.value
	}

	return Header{
		id:    makeHeaderID(name),
		name:  name,
		value: value,
	}
}

// WithValue returns a copy of header with modified value.
func (h *Header) WithValue(value string) Header {
	var (
		name string
		id   string
	)

	if h != nil {
		name = h.name
		id = h.id
	}

	return Header{
		id:    id,
		name:  name,
		value: value,
	}
}

func (h *Header) writeString(writer io.StringWriter) {
	if h == nil {
		writer.WriteString("{no-key}:{no-value}") // nolint: errcheck

		return
	}

	writer.WriteString(h.name)  // nolint: errcheck
	writer.WriteString(":")     // nolint: errcheck
	writer.WriteString(h.value) // nolint: errcheck
}

// NewHeader return a Header instance based on given name and value.
func NewHeader(name, value string) Header {
	return Header{
		id:    makeHeaderID(name),
		name:  name,
		value: value,
	}
}
