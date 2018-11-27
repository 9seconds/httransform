package main

import (
	"net"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

const (
	defaultHTTPPort  = 80
	defaultHTTPSPort = 443
)

func ExtractHost(rawurl string) (string, error) {
	host, _, err := net.SplitHostPort(rawurl)
	if err != nil {
		return "", errors.Annotate(err, "Cannot split host/port")
	}

	return host, nil
}

func MakeBadResponse(resp *fasthttp.Response, msg string, statusCode int) {
	resp.Reset()
	resp.SetBodyString(msg)
	resp.SetStatusCode(statusCode)
	resp.Header.SetContentType("text/plain")
}
