package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/valyala/fasthttp"
)

func ProxyAuthenticate(headerValue, expectedAuth []byte) error {
	pos := bytes.IndexByte(headerValue, ' ')
	if pos < 0 {
		return ErrMalformedHeaderValue
	}

	if !bytes.EqualFold(headerValue[:pos], []byte("Basic")) {
		return fmt.Errorf("unsupported auth schema %s", string(headerValue[:pos]))
	}

	for pos < len(headerValue) && (headerValue[pos] == ' ' || headerValue[pos] == '\t') {
		pos++
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(headerValue[pos:])))
	n, err := base64.StdEncoding.Decode(decoded, headerValue[pos:])
	decoded = decoded[:n]

	if err != nil {
		return fmt.Errorf("incorrectly encoded authorization payload: %w", err)
	}

	compare := subtle.ConstantTimeCompare(decoded[:n], expectedAuth)
	if compare != 1 {
		return ErrFailedAuth
	}

	return nil
}

func AuthenticateRequestHeaders(headers *fasthttp.RequestHeader, expectedAuth []byte) error {
	var authError error

	if len(expectedAuth) == 0 {
		return nil
	}

	headers.VisitAll(func(key, value []byte) {
		if bytes.EqualFold(key, []byte("Proxy-Authorization")) {
			authError = ProxyAuthenticate(value, expectedAuth)
		}
	})

	return authError
}
