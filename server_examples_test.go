package httransform

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

// This is an example of server which runs on some local port, checks
// authorization with user:password credentials.
func ExampleServer() { // nolint: funlen
	// These 2 lines are example keys. If you want to generate your
	// own CA certificate and private key, please run following command:
	//
	// openssl req -x509 -newkey rsa:1024 -keyout private-key.pem -out ca.crt -days 3650 -nodes
	//
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

	ln, err := net.Listen("tcp", "127.0.0.1:0") // 0 forces kernel to choose random port
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

	go srv.Serve(ln) // nolint: errcheck

	proxyURL := &url.URL{
		Host:   ln.Addr().String(),
		Scheme: "http",
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, _ := client.Get("http://httpbin.org/status/200")
	fmt.Println(resp.Status)

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		panic(err)
	}

	resp.Body.Close()

	proxyURL.User = url.UserPassword("user", "password")
	client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	resp, err = client.Get("http://httpbin.org/status/200")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	fmt.Println(resp.Status)
	// Output:
	// 407 Proxy Authentication Required
	// 200 OK
}
