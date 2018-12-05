package httransform

import (
	"bufio"
	"bytes"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/juju/errors"
)

// Header is a structure to present both key and value of the header.
type Header struct {
	// HTTP headers are case-insensitive so we need to have some case
	// insensitive ID to be able to refer them. ID usually stores
	// lowercased key.
	ID string

	// Key stores the name of the header (with correct case).
	Key []byte

	// Value stores value of the header (with correct case).
	Value []byte
}

// String method returns some string representation of this
// datastructure so *Header implements fmt.Stringer interface.
func (h *Header) String() string {
	return fmt.Sprintf("%s: %s", string(h.Key), string(h.Value))
}

// HeaderSet presents collection of headers keeping their order.
//
// If you delete a header, then other headers will keep their positions.
// If you update the value again, then original position is going to be
// restored.
//
//   hs := HeaderSet{}
//   hs.SetString("Connection", "close")
//   hs.GetString("connection")
//
// This will return you "close". Please pay attention to the fact that
// we've used lowercased version. It works. And it works even for updates.
//
//   hs := HeaderSet{}
//   hs.SetString("Connection", "close")
//   hs.SetString("connection", "keep-alive")
//   hs.GetString("Connection")
//
// This will return you "keep-alive". What will you see on the wire?
// Header name will be "Connection" because we've seen such case in the
// first place.
type HeaderSet struct {
	index          map[string]int
	values         []*Header
	removedHeaders map[string]struct{}
}

// SetString sets key and value of the header. If we've already seen
// such header (ignoring its case), we'll update it value, otherwise
// header will be appended to the list. If header was deleted, it would
// be restored keeping the original position in header set.
func (hs *HeaderSet) SetString(key, value string) {
	hs.SetBytes([]byte(key), []byte(value))
}

// SetBytes is the version of SetString which works with bytes.
func (hs *HeaderSet) SetBytes(key []byte, value []byte) {
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

// DeleteString removes (marks as deleted) a header with the given key
// from header set.
func (hs *HeaderSet) DeleteString(key string) {
	hs.removedHeaders[strings.ToLower(key)] = struct{}{}
}

// DeleteBytes is a version of DeleteString which accepts bytes.
func (hs *HeaderSet) DeleteBytes(key []byte) {
	hs.removedHeaders[string(bytes.ToLower(key))] = struct{}{}
}

// GetString returns a value of the header with the given name (ignoring
// the case). If no such header exist, then second parameter is false.
//
// Semantically, this method is similar to accessing map's key.
func (hs *HeaderSet) GetString(key string) (string, bool) {
	lowerKey := strings.ToLower(key)

	_, ok := hs.removedHeaders[lowerKey]
	if ok {
		return "", false
	}

	if pos, ok := hs.index[lowerKey]; ok {
		return string(hs.values[pos].Value), true
	}

	return "", false
}

// GetBytes is the version of GetString which works with bytes.
func (hs *HeaderSet) GetBytes(key []byte) ([]byte, bool) {
	key = bytes.ToLower(key)

	if _, ok := hs.removedHeaders[string(key)]; ok {
		return nil, false
	}
	if pos, ok := hs.index[string(key)]; ok {
		return hs.values[pos].Value, true
	}

	return nil, false
}

// Items returns a list of headers from header set in the correct order.
func (hs *HeaderSet) Items() []*Header {
	headers := make([]*Header, 0, len(hs.values))
	for _, v := range hs.values {
		if _, ok := hs.removedHeaders[v.ID]; !ok {
			headers = append(headers, v)
		}
	}
	return headers
}

// String returns a string representation of the header set. So,
// *HeaderSet implements fmt.Stringer interface.
func (hs *HeaderSet) String() string {
	builder := strings.Builder{}
	for _, v := range hs.Items() {
		builder.WriteString(v.String())
		builder.WriteString("\n")
	}
	return builder.String()
}

func parseHeaders(hset *HeaderSet, rawHeaders []byte) error {
	reader := textproto.NewReader(bufio.NewReader(bytes.NewReader(rawHeaders)))
	// skip first line of HTTP
	reader.ReadContinuedLineBytes() // nolint: errcheck

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

		hset.SetBytes(line[:endKey], line[colPosition:])
	}
}
