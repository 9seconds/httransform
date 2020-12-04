package dialers

import (
	"fmt"
	"net/url"
	"strings"
)

// DialerFromURL construct a dialer based on given URI. You
// can think about this function as a factory to make dialers.
//
// Examples:
//
//   DialerFromURL(opt, "http://user:password@myhost:3561")
//   DialerFromURL(opt, "https://user:password@myhost:3561")
//
// will make dialers which treat myhost as HTTP proxy.
//
//   DialerFromURL(opt, "socks5://user:password@myhost:441")
//
// will make dialer which treat myhost as SOCKS5 proxy.
//
// Please pay attention that host part HAS TO HAVE port. As with
// NewProxyAuth, there is no any implicit guessing.
func DialerFromURL(opt Opts, proxyURL string) (Dialer, error) {
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse proxy url: %w", err)
	}

	username := ""
	password := ""

	if parsed.User != nil {
		username = parsed.User.Username()
		password, _ = parsed.User.Password()
	}

	proxyAuth, err := NewProxyAuth(parsed.Host, username, password)
	if err != nil {
		return nil, fmt.Errorf("incorrect proxy auth credentials: %w", err)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "socks5":
		return NewSocks5(opt, proxyAuth)
	case "http", "https":
		return NewHTTPProxy(opt, proxyAuth), nil
	}

	return nil, fmt.Errorf("unknown proxy scheme: %s", parsed.Scheme)
}
