package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/valyala/fasthttp"
)

type basicAuth struct {
	credentials map[string][]byte
}

func (b *basicAuth) Authenticate(ctx *fasthttp.RequestCtx) (string, error) {
	user := ""
	authError := ErrAuthRequired

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		if bytes.EqualFold(key, []byte("Proxy-Authorization")) {
			user, authError = b.doAuth(value)
		}
	})

	return user, authError
}

func (b *basicAuth) doAuth(header []byte) (string, error) {
	pos := bytes.IndexByte(header, ' ')
	if pos < 0 {
		return "", ErrMalformedHeaderValue
	}

	if !bytes.EqualFold(header[:pos], []byte("Basic")) {
		return "", fmt.Errorf("unsupported auth schema %s", string(header[:pos]))
	}

	for pos < len(header) && (header[pos] == ' ' || header[pos] == '\t') {
		pos++
	}

	toCompare := header[pos:]
	user := ""

	var counter int32

	for k, v := range b.credentials {
		value := int32(subtle.ConstantTimeCompare(toCompare, v))
		counter += value

		if subtle.ConstantTimeEq(value, 1) == 1 {
			user = k
		}
	}

	if subtle.ConstantTimeEq(counter, 1) == 1 {
		return user, nil
	}

	return "", ErrFailedAuth
}

// NewBasicAuth returns an implementation of authenticator which does
// proxy authorization in a basic auth fashion. Please see RFC2617 for
// the reference:
//
// https://tools.ietf.org/html/rfc2617#section-2
//
// Parameter is a map of user to password. Key is the username, password
// is a password.
//
// This authenticator is implemented to work with RequestCtx with no
// normalization.
func NewBasicAuth(credentials map[string]string) Interface {
	processedCredentials := map[string][]byte{}

	for k, v := range credentials {
		userpassword := []byte(k + ":" + v)
		encoded := base64.StdEncoding.EncodeToString(userpassword)

		processedCredentials[k] = []byte(encoded)
	}

	return &basicAuth{
		credentials: processedCredentials,
	}
}
