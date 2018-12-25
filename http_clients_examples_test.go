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
	executor := MakeDefaultSOCKS5ProxyClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

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
	executor := MakeDefaultCONNECTProxyClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Please pay attention that we can access only https:// with this client.
	req.SetRequestURI("https://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}
