package httransform

import (
	"bytes"
	"encoding/base64"
	"net/url"

	"github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

// MakeSimpleResponse is a shortcut function which sets given message
// and status code to response. This is destructive function, all
// existing contents of the response are cleared, even body.
func MakeSimpleResponse(resp *fasthttp.Response, msg string, statusCode int) {
	resp.Reset()
	resp.SetBodyString(msg)
	resp.SetStatusCode(statusCode)
	resp.Header.SetContentType("text/plain")
}

// ExtractAuthentication parses the value of Proxy-Authorization header
// and returns values for user and password. Only Basic authentication
// scheme is supported.
func ExtractAuthentication(text []byte) ([]byte, []byte, error) {
	pos := bytes.IndexByte(text, ' ')
	if pos < 0 {
		return nil, nil, xerrors.New("Malformed Proxy-Authorization header")
	}

	if !bytes.Equal(text[:pos], []byte("Basic")) {
		return nil, nil, xerrors.New("Incorrect authorization prefix")
	}

	for pos < len(text) && (text[pos] == ' ' || text[pos] == '\t') {
		pos++
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(text[pos:])))
	n, err := base64.StdEncoding.Decode(decoded, text[pos:])
	decoded = decoded[:n]

	if err != nil {
		return nil, nil, xerrors.Errorf("incorrectly encoded authorization payload: %w", err)
	}

	pos = bytes.IndexByte(decoded, ':')
	if pos < 0 {
		return nil, nil, xerrors.New("Cannot find a user/password delimiter in decoded authorization string")
	}

	return decoded[:pos], decoded[pos+1:], nil
}

// MakeProxyAuthorizationHeaderValue builds a value of
// Proxy-Authorization header with Basic authentication scheme based on
// information, given in URL.
//
// If no user/pass is defined, then function returns nil.
func MakeProxyAuthorizationHeaderValue(proxyURL *url.URL) []byte {
	username := proxyURL.User.Username()
	password, ok := proxyURL.User.Password()

	if ok || username != "" {
		line := username + ":" + password
		return []byte("Basic " + base64.StdEncoding.EncodeToString([]byte(line)))
	}

	return nil
}
