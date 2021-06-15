package httransform_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/9seconds/httransform/v2"
	"github.com/9seconds/httransform/v2/auth"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/suite"
)

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

type ServerTestSuite struct {
	suite.Suite

	httpEndpoint *httptest.Server
	tlsEndpoint  *httptest.Server
	proxy        *httransform.Server
	ln           net.Listener
	ctx          context.Context
	ctxCancel    context.CancelFunc
	http         *http.Client
}

func (suite *ServerTestSuite) SetupSuite() {
	httpbinApp := httpbin.NewHTTPBin()
	mux := http.NewServeMux()

	mux.HandleFunc("/ip", httpbinApp.IP)

	suite.httpEndpoint = httptest.NewServer(mux)
	suite.tlsEndpoint = httptest.NewTLSServer(mux)
}

func (suite *ServerTestSuite) TearDownSuite() {
	suite.httpEndpoint.Close()
	suite.tlsEndpoint.Close()
}

func (suite *ServerTestSuite) SetupTest() {
	suite.ctx, suite.ctxCancel = context.WithCancel(context.Background())

	opts := httransform.ServerOpts{
		Authenticator: auth.NewBasicAuth(map[string]string{
			"user": "password",
		}),
		TLSCertCA:     caCert,
		TLSPrivateKey: caPrivateKey,
		Layers: []layers.Layer{
			layers.TimeoutLayer{
				Timeout: 10 * time.Second,
			},
		},
	}

	suite.proxy, _ = httransform.NewServer(suite.ctx, opts)
	suite.ln, _ = net.Listen("tcp", "127.0.0.1:0")

	go suite.proxy.Serve(suite.ln)

	httpProxyURL, _ := url.Parse("http://" + suite.ln.Addr().String())
	httpProxyURL.User = url.UserPassword("user", "password")

	suite.http = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(httpProxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second,
	}
}

func (suite *ServerTestSuite) TearDownTest() {
	suite.ctxCancel()
	suite.proxy.Close()
	suite.ln.Close()
}

func (suite *ServerTestSuite) TestIncorrectCACert() {
	opts := httransform.ServerOpts{
		TLSCertCA:     []byte{1, 2, 3},
		TLSPrivateKey: caPrivateKey,
	}

	_, err := httransform.NewServer(suite.ctx, opts)

	suite.Error(err)
}

func (suite *ServerTestSuite) TestIncorrectPrivateKey() {
	opts := httransform.ServerOpts{
		TLSCertCA:     caCert,
		TLSPrivateKey: []byte{1, 2, 3},
	}

	_, err := httransform.NewServer(suite.ctx, opts)

	suite.Error(err)
}

func (suite *ServerTestSuite) TestHTTPRequest() {
	resp, err := suite.http.Get(suite.httpEndpoint.URL + "/ip")

	defer func() {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	suite.NoError(err)

	data, err := ioutil.ReadAll(resp.Body)

	suite.NoError(err)

	v := map[string]interface{}{}

	suite.NoError(json.Unmarshal(data, &v))
}

func (suite *ServerTestSuite) TestHTTPSRequest() {
	resp, err := suite.http.Get(suite.tlsEndpoint.URL + "/ip")

	defer func() {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	suite.NoError(err)

	data, err := ioutil.ReadAll(resp.Body)

	suite.NoError(err)

	v := map[string]interface{}{}

	suite.NoError(json.Unmarshal(data, &v))
}

func (suite *ServerTestSuite) TestHTTPAuthRequired() {
	httpProxyURL, _ := url.Parse("http://" + suite.ln.Addr().String())

	suite.http.Transport.(*http.Transport).Proxy = http.ProxyURL(httpProxyURL)

	resp, err := suite.http.Get(suite.httpEndpoint.URL + "/ip")

	defer func() {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	suite.NoError(err)
	suite.Equal(http.StatusProxyAuthRequired, resp.StatusCode)
}

func (suite *ServerTestSuite) TestHTTPSAuthRequired() {
	httpProxyURL, _ := url.Parse("http://" + suite.ln.Addr().String())

	suite.http.Transport.(*http.Transport).Proxy = http.ProxyURL(httpProxyURL)

	resp, err := suite.http.Get(suite.tlsEndpoint.URL + "/ip")

	defer func() {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	suite.Error(err)
}

func (suite *ServerTestSuite) TestGolangOrg() {
	resp, err := suite.http.Get("https://golang.org")

	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func TestServer(t *testing.T) {
	suite.Run(t, &ServerTestSuite{})
}
