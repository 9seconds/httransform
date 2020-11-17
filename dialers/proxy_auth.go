package dialers

import (
	"fmt"
	"net"
)

type ProxyAuth struct {
	Address  string
	Username string
	Password string
}

func (p *ProxyAuth) HasCredentials() bool {
	return p.Username != "" || p.Password != ""
}

func NewProxyAuth(address, username, password string) (ProxyAuth, error) {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return ProxyAuth{}, fmt.Errorf("incorrect proxy address: %w", err)
	}

	return ProxyAuth{
		Address:  address,
		Username: username,
		Password: password,
	}, nil
}
