package main

import "github.com/juju/errors"

var (
	ProxyAuthorizationError = errors.New("Cannot authenticate proxy user")
)
