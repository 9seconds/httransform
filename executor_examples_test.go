package httransform

import "net/url"

// Let's assume that our SOCKS5 proxy is placed on 127.0.0.1:4040
// address. Then we have to build the correct URL with socks5 scheme.
func ExampleMakeProxyChainExecutor_socks5() {
	socksURL := &url.URL{
		Scheme: "socks5",
		Host:   "127.0.0.1:4040",
	}
	state := &LayerState{}

	executor, _ := MakeProxyChainExecutor(socksURL)
	executor(state)
}

// Let's assume that our HTTP proxy is placed on 127.0.0.1:3128 address.
// Then we have to build the correct URL with HTTP scheme. Also, please
// remember that this executor will support both HTTP and HTTP proxies
// (i.e, will do CONNECT method for HTTPS).
func ExampleMakeProxyChainExecutor_http() {
	httpURL := &url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3128",
	}
	state := &LayerState{}

	executor, _ := MakeProxyChainExecutor(httpURL)
	executor(state)
}
