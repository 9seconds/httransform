httransform
===========

[![Build Status](https://travis-ci.org/9seconds/httransform.svg?branch=master)](https://travis-ci.org/9seconds/httransform)
[![CodeCov](https://codecov.io/gh/9seconds/httransform/branch/master/graph/badge.svg)](https://codecov.io/gh/9seconds/httransform)
[![GoDoc Reference](https://camo.githubusercontent.com/7540274b4c20318e1b1f2d8abe11ba136c926233/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f76616c79616c612f66617374687474703f7374617475732e737667)](https://godoc.org/github.com/9seconds/httransform)
[![Go Report Card](https://goreportcard.com/badge/github.com/9seconds/httransform)](https://goreportcard.com/report/github.com/9seconds/httransform)

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

Please check the [full
documentation](https://godoc.org/github.com/9seconds/httransform) for 
more details.

Example
-------

Just a small example to give you the feeling of how it all looks like:

```go
package main

import (
    "net"

    "github.com/9seconds/httransform"
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
	ln, err := net.Listen("tcp", "127.0.0.1:3128")
	if err != nil {
		panic(err)
	}
	opts := ServerOpts{
		CertCA:  caCert,
		CertKey: caPrivateKey,
		Layers: []Layer{
			&ProxyAuthorizationBasicLayer{
				User:     []byte("user"),
				Password: []byte("password"),
				Realm:    "test",
			},
			&AddRemoveHeaderLayer{
				AbsentRequestHeaders: []string{"proxy-authorization"},
			},
		},
	}
	srv, err := NewServer(opts)
	if err != nil {
		panic(err)
	}

	if err := srv.Serve(ln); err != nil {
		panic(err)
	}
}
```

This will create an HTTP proxy on `127.0.0.1:3128`. It will also require
authentication (`user` and `password`) and will remove the `Proxy-Authorization`
header before sending the request further.
