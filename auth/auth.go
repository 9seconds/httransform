package auth

import (
	"bytes"
	"crypto/subtle"
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

	if subtle.ConstantTimeCompare(headerValue[pos:], expectedAuth) != 1 {
		return ErrFailedAuth
	}

	return nil
}

func AuthenticateRequestHeaders(headers *fasthttp.RequestHeader, expectedAuth []byte) error {
	if len(expectedAuth) == 0 {
		return nil
	}

	authError := ErrFailedAuth

	headers.VisitAll(func(key, value []byte) {
		if bytes.EqualFold(key, []byte("Proxy-Authorization")) {
			authError = ProxyAuthenticate(value, expectedAuth)
		}
	})

	return authError
}
