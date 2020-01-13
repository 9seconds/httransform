package utils

import (
	"bytes"
	"encoding/base64"
	"strings"
)

func ExtractAuthentication(text string) (string, string, error) {
	pos := strings.IndexByte(text, ' ')
	if pos < 0 {
		return "", "", ErrExtractAuthenticationMalformed
	}

	if !strings.EqualFold(text[:pos], "Basic") {
		return "", "", ErrExtractAuthenticationPrefix
	}

	for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
		pos++
	}

	decoded, err := base64.StdEncoding.DecodeString(text[pos:])
	if err != nil {
		return "", "", ErrExtractAuthentication.Wrap("incorrect encoded payload", err)
	}

	pos = bytes.IndexByte(decoded, ':')
	if pos < 0 {
		return "", "", ErrExtractAuthenticationDelimiter
	}

	return string(decoded[:pos]), string(decoded[pos+1:]), nil
}
