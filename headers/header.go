package headers

import (
	"io"
	"net/textproto"
	"strings"
)

type Header struct {
	id    []byte
	value string
	name  string
}

func (h *Header) String() string {
	builder := strings.Builder{}

	h.writeString(&builder)

	return builder.String()
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

func (h *Header) ID() []byte {
	if h == nil {
		return nil
	}

	return h.id
}

func (h *Header) Name() string {
	if h == nil {
		return ""
	}

	return h.name
}

func (h *Header) CanonicalName() string {
	return textproto.CanonicalMIMEHeaderKey(h.Name())
}

func (h *Header) Value() string {
	if h == nil {
		return ""
	}

	return h.value
}

func (h *Header) Values() []string {
	return Values(h.Value())
}

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

func (h *Header) WithValue(value string) Header {
	var (
		name string
		id   []byte
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

func NewHeader(name, value string) Header {
	return Header{
		id:    makeHeaderID(name),
		name:  name,
		value: value,
	}
}
