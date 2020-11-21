package headers

import (
	"bufio"
	"bytes"
	"net/textproto"
	"strings"
)

func makeHeaderID(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func makeTextProtoReader(data []byte) *textproto.Reader {
	rawReader := bytes.NewReader(data)
	bufioReader := bufio.NewReader(rawReader)
	textProtoReader := textproto.NewReader(bufioReader)

	return textProtoReader
}
