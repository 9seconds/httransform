package main

import "fmt"

type Header struct {
	ID    string
	Key   string
	Value []byte
}

func (h *Header) String() string {
	return fmt.Sprintf("%s: %s", h.Key, string(h.Value))
}
