package headers

import (
	"net/textproto"
)

type Header struct {
	id    []byte
	value string
	name  string
}

func (h *Header) String() string {
	name := h.Name()
	if name == "" {
		name = "{empty}"
	}

	value := h.Value()
	if value == "" {
		value = "{empty}"
	}

	return name + ":" + value
}

func (h *Header) ID() []byte {
	if h == nil {
		return nil
	}

	if h.id == nil {
		h.id = makeHeaderID(h.Name())
	}

	return h.id
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

func (h *Header) Name() string {
	if h == nil {
		return ""
	}

	return h.name
}

func (h *Header) CanonicalName() string {
	return textproto.CanonicalMIMEHeaderKey(h.Name())
}

func NewHeader(name, value string) Header {
	return Header{
		id:    makeHeaderID(name),
		name:  name,
		value: value,
	}
}
