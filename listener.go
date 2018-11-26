package main

import (
	"net"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp/reuseport"
)

func MakeListener(network, addr string, forceSoReusePort bool) (net.Listener, error) {
	ln, err := reuseport.Listen(network, addr)
	if err != nil {
		if forceSoReusePort {
			return nil, errors.Annotate(err, "Cannot create SO_REUSEPORT listener")
		}
		return net.Listen(network, addr)
	}

	return ln, nil
}
