httransform
===========

[![Build Status](https://github.com/9seconds/httransform/workflows/CI/badge.svg)](https://github.com/9seconds/httransform)
[![codecov](https://codecov.io/gh/9seconds/httransform/branch/master/graph/badge.svg?token=cyMF66trUZ)](https://codecov.io/gh/9seconds/httransform)
[![Go Reference](https://pkg.go.dev/badge/github.com/9seconds/httransform/v2.svg)](https://pkg.go.dev/github.com/9seconds/httransform/v2)

httransform is the library/framework to build your own HTTP
proxies. It relies on high-performant and memory-efficient
[fasthttp](https://github.com/valyala/fasthttp) library as HTTP base
layer and can give you the ability to build a proxy where you can control
every aspect.

Main features of this framework:

1. Support of HTTP (plain HTTP) proxy protocol.
2. Support of HTTPS (with CONNECT method) protocol. This library does MITM
   and provides the possibility to generate TLS certificates for the hosts
   on-the-fly.
3. Keeps and maintains the order of headers and their case (no normalization).
4. Supports the concept of _layers_ or middlewares which process HTTP
   requests and responses.
5. Supports custom _executors_: a function which converts HTTP requests to
   HTTP responses. Allowing your proxy to fetch the data from other services,
   which are not necessarily HTTP. The executor simply converts HTTP request
   structure to HTTP response.
6. Can support connection upgrades (this includes _websockets_) with an
   ability to look into them.

Please check the [full
documentation](https://pkg.go.dev/github.com/9seconds/httransform/v2) for
more details.

Example
-------

Just a small example to give you the feeling of how it all looks like:

```go
package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/9seconds/httransform/v2"
	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/layers"
)

// These are generates examples of self-signed certificates
// to simplify the example.
var caCert = []byte(`-----BEGIN CERTIFICATE-----
MIICWzCCAcSgAwIBAgIJAJ34yk7oiKv5MA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTgxMjAyMTQyNTAyWhcNMjgxMTI5MTQyNTAyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKB
gQDL7Hzfmx7xfFWTRm26t/lLsCZwOri6VIzp2dYM5Hp0dV4XUZ+q60nEbHwN3Usr
GKAK/Rsr9Caam3A18Upn2ly69Tyr29kVK+PlsOgSSCUnAYcqT166/j205n3CGNLL
OPtQKfAT/iH3dPBObd8N4FR9FlXiYIiAp1opCbyu2mlHiwIDAQABo1MwUTAdBgNV
HQ4EFgQUOJ+uGtIhHxXHPNESBNI4YbwAl+wwHwYDVR0jBBgwFoAUOJ+uGtIhHxXH
PNESBNI4YbwAl+wwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQCW
s7P0wJ8ON8ieEJe4pAfACpL6IyhZ5YK/C/hip+czxdvZHc5zngVwHP2vsIcHKBTr
8qXoHgh2gaXqwn8kRVNnZzWrxgSe8IR3oJ2yTbLAxqDS42SPfRLAUpy9sK/tEEGM
rMk/LWMzH/S6bLcsAm0GfVIrUNfg0eF0ZVIjxINBVA==
-----END CERTIFICATE-----`)

var caPrivateKey = []byte(`-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAMvsfN+bHvF8VZNG
bbq3+UuwJnA6uLpUjOnZ1gzkenR1XhdRn6rrScRsfA3dSysYoAr9Gyv0JpqbcDXx
SmfaXLr1PKvb2RUr4+Ww6BJIJScBhypPXrr+PbTmfcIY0ss4+1Ap8BP+Ifd08E5t
3w3gVH0WVeJgiICnWikJvK7aaUeLAgMBAAECgYAk+/kR3OJZzcD/evB/wsoV7haq
mBvUv2znJLjrkayb3oV4GTeqGg5A76P4J8BwSoEMPSdma1ttAu/w+JgUCchzVPwU
34Sr80mYawOmGVGJsDnrrYA2w51Nj42e71pmRc9IqNLwFEhW5Uy7eASf3THJMWDl
F2M6xAVYr+X0eKLf4QJBAO8lVIIMnzIReSZukWBPp6GKmXOuEkWeBOfnYC2HOVZq
1M/E6naOP2MBk9CWG4o9ysjcZ1hosi3/txxrc8VmBAkCQQDaS651dpQ3TRE//raZ
s79ZBEdMCMlgXB6CPrZpvLz/3ZPcLih4MJ59oVkeFHCNct7ccQcQu4XHMGNBIRBh
kpvzAkEAlS/AjHC7T0y/O052upJ2jLweBqBtHaj6foFE6qIVDugOYp8BdXw/5s+x
GsrJ22+49Z0pi2mk3jVMUhpmWprNoQJBANdAT0v2XFpXfQ38bTQMYT82j9Myytdg
npjRm++Rs1AdvoIbZb52OqIoqoaVoxJnVchLD6t5LYXnecesAcok1e8CQEKB7ycJ
6yVwnBE3Ua9CHcGmrre6HmEWdPy1Zyb5DQC6duX46zEBzti9oWx0DJIQRZifeCvw
4J45NsSQjuuAAWs=
-----END PRIVATE KEY-----`)

func main() {
	// Root context is crucial here. When root context is closed, a
	// proxy is shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// For demo purpose we are going to close by SIGINT and SIGTERM
	// signals.
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for range signals {
			cancel()
		}
	}()

	// Filter layer is required if you want to drop request
	// which are made to you internal networks.
	filterLayer, err := layers.NewFilterSubnetsLayer([]net.IPNet{
		{IP: net.ParseIP("127.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	})
	if err != nil {
		panic(err)
	}

	opts := httransform.ServerOpts{
		TLSCertCA:     caCert,
		TLSPrivateKey: caPrivateKey,
		// We are going to use basic proxy authorization with username
		// 'user' and password 'password'.
		Authenticator: auth.NewBasicAuth(map[string]string{
			"user": "password",
		}),
		Layers: []layers.Layer{
			filterLayer,
			// This guy will remove Proxy headers from the request.
			layers.ProxyHeadersLayer{},
			// This guy is going to limit request processing time to 3
			// minutes.
			layers.TimeoutLayer{
				Timeout: 3 * time.Minute,
			},
		},
	}

	proxy, err := httransform.NewServer(ctx, opts)
	if err != nil {
		panic(err)
	}

	// We bind our proxy to the port 3128 and all interfaces.
	listener, err := net.Listen("tcp", ":3128")
	if err != nil {
		panic(err)
	}

	if err := proxy.Serve(listener); err != nil {
		panic(err)
	}
}
```

This will create an HTTP proxy on `127.0.0.1:3128`. It will also require
authentication (`user` and `password`) and will remove the `Proxy-Authorization`
header before sending the request further.


v1 version
----------

Version 1 is not supported anymore. Please use version 2. Version 2 has
a lot of breaking changes but which really help to maintain a package.

Unfortunately, I cannot enumerate all of them and I understand that it
is going to make your life painful in some moments. Sorry for that. v2
was a complete rewrite and I assume that given a size of this library,
it won't take a lot of time for you to migrate. But if you have any
questions, feel free to open an issue. I'm happy to help.
