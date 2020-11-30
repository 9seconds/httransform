package dialers

import (
	"fmt"
	"net/url"
	"strings"
)

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
		return NewSocks(opt, proxyAuth)
	case "http", "https":
		return NewHTTPProxy(opt, proxyAuth), nil
	}

	return nil, fmt.Errorf("unknown proxy scheme: %s", parsed.Scheme)
}
