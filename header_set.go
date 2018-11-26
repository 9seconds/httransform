package main

import (
	"bufio"
	"bytes"
	"net/textproto"
	"strings"

	"github.com/juju/errors"
)

type HeaderSet struct {
	index          map[string]int
	values         []*Header
	removedHeaders map[string]struct{}
}

func (hs *HeaderSet) AddString(key, value string) {
	hs.AddBytes([]byte(key), []byte(value))
}

func (hs *HeaderSet) AddBytes(key []byte, value []byte) {
	lowerKey := string(bytes.ToLower(key))

	if position, ok := hs.index[lowerKey]; ok {
		hs.values[position].Value = value
	} else {
		newHeader := getHeader()
		newHeader.ID = lowerKey
		newHeader.Key = append(newHeader.Key, key...)
		newHeader.Value = append(newHeader.Value, value...)
		hs.values = append(hs.values, newHeader)
		hs.index[lowerKey] = len(hs.values) - 1
	}

	delete(hs.removedHeaders, lowerKey)
}

func (hs *HeaderSet) DeleteString(key string) {
	hs.removedHeaders[strings.ToLower(key)] = struct{}{}
}

func (hs *HeaderSet) DeleteBytes(key []byte) {
	hs.removedHeaders[string(bytes.ToLower(key))] = struct{}{}
}

func (hs *HeaderSet) GetString(key string) (string, bool) {
	lowerKey := strings.ToLower(key)

	_, ok := hs.removedHeaders[lowerKey]
	if !ok {
		return "", false
	}

	if pos, ok := hs.index[lowerKey]; ok {
		return string(hs.values[pos].Value), true
	}

	return "", false
}

func (hs *HeaderSet) GetBytes(key []byte) ([]byte, bool) {
	lowerKey := string(bytes.ToLower(key))

	_, ok := hs.removedHeaders[lowerKey]
	if !ok {
		return nil, false
	}

	if pos, ok := hs.index[lowerKey]; ok {
		return hs.values[pos].Value, true
	}

	return nil, false
}

func (hs *HeaderSet) Items() []*Header {
	headers := make([]*Header, 0, len(hs.values))
	for _, v := range hs.values {
		if _, ok := hs.removedHeaders[v.ID]; !ok {
			headers = append(headers, v)
		}
	}
	return headers
}

func (hs *HeaderSet) Clear() {
	for k := range hs.index {
		delete(hs.index, k)
	}
	for k := range hs.removedHeaders {
		delete(hs.removedHeaders, k)
	}
	for _, v := range hs.values {
		releaseHeader(v)
	}
	hs.values = hs.values[:0]
}

func parseHeaders(hset *HeaderSet, rawHeaders []byte) error {
	reader := textproto.NewReader(bufio.NewReader(bytes.NewReader(rawHeaders)))
	reader.ReadContinuedLineBytes() // skip first line of HTTP

	for {
		line, err := reader.ReadContinuedLineBytes()
		if len(line) == 0 || err != nil {
			return err
		}

		colPosition := bytes.IndexByte(line, ':')
		if colPosition < 0 {
			return errors.Errorf("Malformed header %s", string(line))
		}

		endKey := colPosition
		for endKey > 0 && line[endKey-1] == ' ' {
			endKey--
		}

		colPosition++
		for colPosition < len(line) && (line[colPosition] == ' ' || line[colPosition] == '\t') {
			colPosition++
		}

		hset.AddBytes(line[:endKey], line[colPosition:])
	}
}
