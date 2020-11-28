package headers

type Header struct {
	ID    string
	Name  string
	Value string
}

func (h *Header) GetValue() string {
	if h == nil {
		return ""
	}

	return h.Value
}
