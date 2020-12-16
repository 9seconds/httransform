package dialers

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ProxyAuth contains an information how to authenticate to a given
// proxy.
type ProxyAuth struct {
	// Address is address to a proxy in a form of host:port.
	Address string

	// Username is a name of user used for authentication.
	Username string

	// Password is a password of user used for authentication.
	Password string
}

// HasCredentials checks if this proxy has authentication or not.
func (p *ProxyAuth) HasCredentials() bool {
	return p.Username != "" || p.Password != ""
}

// Host returns a host part of Address.
func (p *ProxyAuth) Host() string {
	host, _, _ := net.SplitHostPort(p.Address)

	return host
}

// Port return a port part of Address.
func (p *ProxyAuth) Port() int {
	_, port, _ := net.SplitHostPort(p.Address)

	number, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	return number
}

// String conforms fmt.Stringer interface.
func (p *ProxyAuth) String() string {
	u := url.URL{
		Host: p.Address,
	}

	if p.HasCredentials() {
		u.User = url.UserPassword(p.Username, p.Password)
	}

	return strings.TrimPrefix(u.String(), "//")
}

// NewProxyAuth return a new ProxyAuth instance according to given
// parameters.
func NewProxyAuth(address, username, password string) (ProxyAuth, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return ProxyAuth{}, fmt.Errorf("incorrect proxy address: %w", err)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return ProxyAuth{
		Address:  net.JoinHostPort(host, port),
		Username: username,
		Password: password,
	}, nil
}
