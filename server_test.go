package httransform

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

var testServerCACert = []byte(`-----BEGIN CERTIFICATE-----
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

var testServerPrivateKey = []byte(`-----BEGIN PRIVATE KEY-----
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

func testServerExecutor(state *LayerState) {
	state.Response.SetStatusCode(fasthttp.StatusNotFound)
	state.Response.SetBodyString("Not found!")
}

type Handler struct {
	callback func(http.ResponseWriter, *http.Request)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.callback(w, r)
}

type MockLayer struct {
	mock.Mock
}

func (m *MockLayer) OnRequest(state *LayerState) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockLayer) OnResponse(state *LayerState, err error) {
	args := m.Called(state, err)
	state.ResponseHeaders.SetString("X-Test", args.Get(0).(string))
}

type BaseServerTestSuite struct {
	suite.Suite

	ln     net.Listener
	client *http.Client
	opts   ServerOpts
}

func (suite *BaseServerTestSuite) SetupTest() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	suite.ln = ln
	suite.opts = ServerOpts{
		CertCA:  testServerCACert,
		CertKey: testServerPrivateKey,
	}

	proxyURL := &url.URL{
		Host:   ln.Addr().String(),
		Scheme: "http",
	}
	suite.client = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint: gosec
		},
	}
}

func (suite *BaseServerTestSuite) TearDownTest() {
	suite.ln.Close()
}

type ServerTestSuite struct {
	BaseServerTestSuite

	srv *Server
}

func (suite *ServerTestSuite) SetupTest() {
	suite.BaseServerTestSuite.SetupTest()

	srv, err := NewServer(suite.opts, []Layer{}, testServerExecutor, &NoopLogger{})
	if err != nil {
		panic(err)
	}
	suite.srv = srv

	go srv.Serve(suite.ln) // nolint: errcheck
}

func (suite *ServerTestSuite) TestHTTPRequest() {
	resp, err := suite.client.Get("http://example.com")

	suite.Equal(resp.StatusCode, http.StatusNotFound)
	suite.Nil(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal(body, []byte("Not found!"))
}

func (suite *ServerTestSuite) TestHTTPSRequest() {
	resp, err := suite.client.Get("https://example.com")

	suite.Equal(resp.StatusCode, http.StatusNotFound)
	suite.Nil(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal(body, []byte("Not found!"))
}

func (suite *ServerTestSuite) TestLayerNoError() {
	mocked := &MockLayer{}
	mocked.On("OnRequest", mock.Anything).Return(nil)
	mocked.On("OnResponse", mock.Anything, nil).Return("value")

	suite.srv.layers = append(suite.srv.layers, mocked)

	resp, err := suite.client.Get("http://example.com")

	suite.Equal(resp.StatusCode, http.StatusNotFound)
	suite.Nil(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal(body, []byte("Not found!"))

	suite.Equal(resp.Header.Get("x-test"), "value")
}

func (suite *ServerTestSuite) TestLayerError() {
	err := errors.New("Some error")
	mocked := &MockLayer{}
	mocked.On("OnRequest", mock.Anything).Return(err)
	mocked.On("OnResponse", mock.Anything, err).Return("value")

	suite.srv.layers = append(suite.srv.layers, mocked)

	resp, err := suite.client.Get("http://example.com")
	suite.Nil(err)

	suite.Equal(resp.StatusCode, http.StatusInternalServerError)

	suite.Equal(resp.Header.Get("x-test"), "value")
}

type ServerProxyChainTestSuite struct {
	BaseServerTestSuite

	endListener net.Listener
}

func (suite *ServerProxyChainTestSuite) SetupTest() {
	suite.BaseServerTestSuite.SetupTest()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	suite.endListener = ln
}

func (suite *ServerProxyChainTestSuite) TearDownTest() {
	suite.BaseServerTestSuite.TearDownTest()

	suite.endListener.Close()
}

func (suite *ServerProxyChainTestSuite) TestChainDropsConnect() {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   suite.endListener.Addr().String(),
	}
	executor, err := MakeProxyChainExecutor(proxyURL)

	suite.Nil(err)

	srv, err := NewServer(suite.opts, []Layer{}, executor, &NoopLogger{})
	suite.Nil(err)

	go srv.Serve(suite.ln)

	called := false
	endSrv := http.Server{
		Handler: &Handler{
			callback: func(w http.ResponseWriter, req *http.Request) {
				if called {
					called = false
					w.WriteHeader(http.StatusNotFound)
				} else {
					called = true
					w.WriteHeader(http.StatusProxyAuthRequired)
				}
			},
		},
	}
	go endSrv.Serve(suite.endListener)

	resp, err := suite.client.Get("http://example.com")
	suite.Nil(err)
	suite.True(called)
	suite.Equal(resp.StatusCode, http.StatusProxyAuthRequired)

	resp, err = suite.client.Get("http://example.com")
	suite.Nil(err)
	suite.False(called)
	suite.Equal(resp.StatusCode, http.StatusNotFound)
}

func TestServer(t *testing.T) {
	suite.Run(t, &ServerTestSuite{})
}

func TestServerProxyChain(t *testing.T) {
	suite.Run(t, &ServerProxyChainTestSuite{})
}
