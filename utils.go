package main

import (
	"net"
	"net/url"
	"strconv"

	"github.com/juju/errors"
	"github.com/valyala/fasthttp"
)

const (
	defaultHTTPPort  = 80
	defaultHTTPSPort = 443
)

func SplitHostPort(rawurl string) (string, int, error) {
	host, port, err := net.SplitHostPort(rawurl)
	if err != nil {
		return "", 0, errors.Annotate(err, "Cannot split host/port")
	}

	portInt, err := parsePort(port)
	if err != nil {
		return "", 0, errors.Annotate(err, "Cannot parse port")
	}

	return host, portInt, nil
}

func HostPortFromURL(rawurl string) (string, int, error) {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return "", 0, errors.Annotate(err, "Cannot parse URL")
	}

	host, port, err := net.SplitHostPort(parsedURL.Host)
	var portInt int
	if err != nil {
		host = parsedURL.Host
		switch parsedURL.Scheme {
		case "http":
			portInt = defaultHTTPPort
		case "https":
			portInt = defaultHTTPSPort
		default:
			return "", 0, errors.Errorf("Unknown default port for scheme %s", parsedURL.Scheme)
		}
	}

	if portInt == 0 {
		portInt, err = parsePort(port)
		if err != nil {
			return "", 0, errors.Annotate(err, "Cannot parse port")
		}
	}

	return host, portInt, nil
}

func parsePort(port string) (int, error) {
	portUint, err := strconv.ParseUint(port, 10, 16)
	return int(portUint), err
}

func MakeBadResponse(resp *fasthttp.Response, msg string, statusCode int) {
	resp.Reset()
	resp.SetBodyString(msg)
	resp.SetStatusCode(statusCode)
	resp.Header.SetContentType("text/plain")
}
