package main

import "fmt"

type Header struct {
	ID    string
	Key   []byte
	Value []byte
}

func (h *Header) String() string {
	return fmt.Sprintf("%s: %s", string(h.Key), string(h.Value))
}
