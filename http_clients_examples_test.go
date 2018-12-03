package httransform

import (
	"net/url"

	"github.com/valyala/fasthttp"
)

// Let's assume that SOCKS5 proxy is placed on "127.0.0.1:4040".
func ExampleMakeProxySOCKS5Client() {
	proxyURL := &url.URL{
		Scheme: "socks5",
		Host:   "127.0.0.1:4040",
	}
	executor, _ := MakeProxySOCKS5Client(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that HTTP proxy is placed on "127.0.0.1:3128".
func ExampleMakeHTTPProxyClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeHTTPProxyClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Please pay attention that we can access only http:// with this client.
	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that HTTPS proxy is placed on "127.0.0.1:3128".
func ExampleMakeHTTPSProxyClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeHTTPSProxyClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Please pay attention that we can access only https:// with this client.
	req.SetRequestURI("https://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}
