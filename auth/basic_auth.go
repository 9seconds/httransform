package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"

	"github.com/valyala/fasthttp"
)

const (
	basicAuthNotFound = -2
	basicAuthFailed   = -1
)

type basicAuthPair struct {
	user  string
	value []byte
}

type basicAuth struct {
	credentials []basicAuthPair
}

func (b *basicAuth) Authenticate(ctx *fasthttp.RequestCtx) (string, error) {
	idx := basicAuthNotFound

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		if bytes.EqualFold(key, []byte("Proxy-Authorization")) {
			idx = b.doAuth(value)
		}
	})

	switch idx {
	case basicAuthNotFound:
		return "", ErrAuthRequired
	case basicAuthFailed:
		return "", ErrFailedAuth
	default:
		return b.credentials[idx].user, nil
	}
}

func (b *basicAuth) doAuth(header []byte) int {
	pos := bytes.IndexByte(header, ' ')
	if pos < 0 {
		return basicAuthFailed
	}

	if !bytes.EqualFold(header[:pos], []byte("Basic")) {
		return basicAuthFailed
	}

	for pos < len(header) && (header[pos] == ' ' || header[pos] == '\t') {
		pos++
	}

	toCompare := header[pos:]
	rv := basicAuthFailed

	// Yes, it looks quite weird and simple basic auth implementation
	// can be very-very simple but if we want to avoid timing attacks,
	// we have to make it constant time. So, this is a reason why we
	// compare and choose with subtle module. And this is a reason why
	// we even find a user with ConstantTimeSelect which is a twist to
	// understand.
	//
	// But idea is simple: negative ints mean failed auth, positive ones
	// - indexes within an array.
	for i := range b.credentials {
		compared := subtle.ConstantTimeCompare(toCompare, b.credentials[i].value)
		rv = subtle.ConstantTimeSelect(compared, i, rv)
	}

	return rv
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
	pairs := make([]basicAuthPair, 0, len(credentials))

	for k, v := range credentials {
		userpassword := []byte(k + ":" + v)
		encoded := base64.StdEncoding.EncodeToString(userpassword)

		pairs = append(pairs, basicAuthPair{
			user:  k,
			value: []byte(encoded),
		})
	}

	return &basicAuth{
		credentials: pairs,
	}
}
