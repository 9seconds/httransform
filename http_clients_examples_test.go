package httransform

import (
	"net/url"

	"github.com/valyala/fasthttp"
)

func ExampleMakeStreamingClosingHTTPClient() {
	executor := MakeStreamingClosingHTTPClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that SOCKS5 proxy is placed on "127.0.0.1:4040".
func ExampleMakeStreamingClosingSOCKS5HTTPClient() {
	proxyURL := &url.URL{
		Scheme: "socks5",
		Host:   "127.0.0.1:4040",
	}
	executor := MakeStreamingClosingSOCKS5HTTPClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that HTTP proxy is placed on "127.0.0.1:3128".
func ExampleMakeStreamingClosingProxyHTTPClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeStreamingClosingProxyHTTPClient(proxyURL)

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
func ExampleMakeStreamingClosingCONNECTHTTPClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeStreamingClosingCONNECTHTTPClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Please pay attention that we can access only https:// with this client.
	req.SetRequestURI("https://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

func ExampleMakeStreamingReuseHTTPClient() {
	executor := MakeStreamingReuseHTTPClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that SOCKS5 proxy is placed on "127.0.0.1:4040".
func ExampleMakeStreamingReuseSOCKS5HTTPClient() {
	proxyURL := &url.URL{
		Scheme: "socks5",
		Host:   "127.0.0.1:4040",
	}
	executor := MakeStreamingReuseSOCKS5HTTPClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that HTTP proxy is placed on "127.0.0.1:3128".
func ExampleMakeStreamingReuseProxyHTTPClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeStreamingReuseProxyHTTPClient(proxyURL)

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
func ExampleMakeStreamingReuseCONNECTHTTPClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeStreamingReuseCONNECTHTTPClient(proxyURL)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Please pay attention that we can access only https:// with this client.
	req.SetRequestURI("https://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

func ExampleMakeDefaultHTTPClient() {
	executor := MakeDefaultHTTPClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("http://google.com")
	req.Header.SetMethod("GET")

	_ = executor.Do(req, resp)
}

// Let's assume that SOCKS5 proxy is placed on "127.0.0.1:4040".
func ExampleMakeDefaultSOCKS5ProxyClient() {
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

// Let's assume that HTTP proxy is placed on "127.0.0.1:3128".
func ExampleMakeDefaultHTTPProxyClient() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	executor := MakeDefaultHTTPProxyClient(proxyURL)

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
func ExampleMakeDefaultCONNECTProxyClient() {
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
