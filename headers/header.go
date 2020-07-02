package headers

import "fmt"

type Header struct {
	id string

	Name  string
	Value string
}

func (h *Header) String() string {
	return fmt.Sprintf("%s: %s", h.Name, h.Value)
}

func (h *Header) Reset() {
	h.id = ""
	h.Name = ""
	h.Value = ""
}
