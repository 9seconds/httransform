package httransform

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
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

type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) NewConnection()   { m.Called() }
func (m *MockMetrics) DropConnection()  { m.Called() }
func (m *MockMetrics) NewGet()          { m.Called() }
func (m *MockMetrics) NewHead()         { m.Called() }
func (m *MockMetrics) NewPost()         { m.Called() }
func (m *MockMetrics) NewPut()          { m.Called() }
func (m *MockMetrics) NewDelete()       { m.Called() }
func (m *MockMetrics) NewConnect()      { m.Called() }
func (m *MockMetrics) NewOptions()      { m.Called() }
func (m *MockMetrics) NewTrace()        { m.Called() }
func (m *MockMetrics) NewPatch()        { m.Called() }
func (m *MockMetrics) NewOther()        { m.Called() }
func (m *MockMetrics) DropGet()         { m.Called() }
func (m *MockMetrics) DropHead()        { m.Called() }
func (m *MockMetrics) DropPost()        { m.Called() }
func (m *MockMetrics) DropPut()         { m.Called() }
func (m *MockMetrics) DropDelete()      { m.Called() }
func (m *MockMetrics) DropConnect()     { m.Called() }
func (m *MockMetrics) DropOptions()     { m.Called() }
func (m *MockMetrics) DropTrace()       { m.Called() }
func (m *MockMetrics) DropPatch()       { m.Called() }
func (m *MockMetrics) DropOther()       { m.Called() }
func (m *MockMetrics) NewCertificate()  { m.Called() }
func (m *MockMetrics) DropCertificate() { m.Called() }

type BaseServerTestSuite struct {
	suite.Suite

	ln      net.Listener
	client  *http.Client
	opts    ServerOpts
	metrics *MockMetrics
}

func (suite *BaseServerTestSuite) SetupTest() {
	suite.metrics = &MockMetrics{}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	suite.ln = ln
	suite.opts = ServerOpts{
		CertCA:   testServerCACert,
		CertKey:  testServerPrivateKey,
		Executor: testServerExecutor,
		Metrics:  suite.metrics,
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

	srv, err := NewServer(suite.opts)
	if err != nil {
		panic(err)
	}
	suite.srv = srv

	go srv.Serve(suite.ln) // nolint: errcheck
}

func (suite *ServerTestSuite) TestHTTPRequest() {
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

	resp, err := suite.client.Get("http://example.com")

	suite.Equal(resp.StatusCode, http.StatusNotFound)
	suite.Nil(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal(body, []byte("Not found!"))
}

func (suite *ServerTestSuite) TestHTTPSRequest() {
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewGet")
	suite.metrics.On("NewConnect")
	suite.metrics.On("NewCertificate")
	suite.metrics.On("DropGet")
	suite.metrics.On("DropConnect")
	suite.metrics.On("DropCertificate")

	resp, err := suite.client.Get("https://example.com")

	suite.Equal(resp.StatusCode, http.StatusNotFound)
	suite.Nil(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal(body, []byte("Not found!"))
}

func (suite *ServerTestSuite) TestLayerNoError() {
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

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
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

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

	status      int
	endSrv      *http.Server
	endListener net.Listener
}

func (suite *ServerProxyChainTestSuite) SetupTest() {
	suite.BaseServerTestSuite.SetupTest()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	suite.endListener = ln

	proxyURL := &url.URL{
		Scheme: "http",
		Host:   suite.endListener.Addr().String(),
	}
	executor, err := MakeProxyChainExecutor(proxyURL)
	suite.Nil(err)
	suite.opts.Executor = executor

	srv, err := NewServer(suite.opts)
	suite.Nil(err)

	go srv.Serve(suite.ln) // nolint: errcheck

	suite.status = http.StatusOK
	suite.endSrv = &http.Server{
		Handler: &Handler{
			callback: func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(suite.status)
			},
		},
	}
}

func (suite *ServerProxyChainTestSuite) TearDownTest() {
	suite.BaseServerTestSuite.TearDownTest()

	suite.endListener.Close()
}

func (suite *ServerProxyChainTestSuite) TestChainDropsConnectOnHTTP() {
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

	go suite.endSrv.Serve(suite.endListener) // nolint: errcheck

	suite.status = http.StatusProxyAuthRequired
	resp, err := suite.client.Get("http://example.com")
	suite.Nil(err)
	suite.Equal(resp.StatusCode, http.StatusProxyAuthRequired)

	suite.status = http.StatusNotFound
	resp, err = suite.client.Get("http://example.com")
	suite.Nil(err)
	suite.Equal(resp.StatusCode, http.StatusNotFound)
}

func (suite *ServerProxyChainTestSuite) TestChainDropsConnectOnHTTPSErrors() {
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewConnect")
	suite.metrics.On("DropConnect")
	suite.metrics.On("NewCertificate")
	suite.metrics.On("DropCertificate")
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

	go suite.endSrv.Serve(suite.endListener) // nolint: errcheck

	suite.status = http.StatusProxyAuthRequired
	resp, err := suite.client.Get("https://example.com")
	suite.Nil(err)
	suite.Equal(resp.StatusCode, http.StatusBadGateway)

	suite.status = http.StatusOK
	resp, err = suite.client.Get("https://example.com")
	suite.Nil(err)
	suite.Equal(resp.StatusCode, http.StatusBadGateway) // TLS response
}

func (suite *ServerProxyChainTestSuite) TestChainConnectOnHTTPS() {
	suite.metrics.On("NewConnection")
	suite.metrics.On("DropConnection")
	suite.metrics.On("NewConnect")
	suite.metrics.On("DropConnect")
	suite.metrics.On("NewCertificate")
	suite.metrics.On("DropCertificate")
	suite.metrics.On("NewGet")
	suite.metrics.On("DropGet")

	certFile, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(certFile.Name())

	if _, err = certFile.Write(testServerCACert); err != nil {
		panic(err)
	}
	if err = certFile.Sync(); err != nil {
		panic(err)
	}

	certKey, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(certKey.Name())

	if _, err = certKey.Write(testServerPrivateKey); err != nil {
		panic(err)
	}
	if err = certKey.Sync(); err != nil {
		panic(err)
	}

	go suite.endSrv.ServeTLS(suite.endListener, certFile.Name(), certKey.Name()) // nolint: errcheck

	suite.status = http.StatusProxyAuthRequired
	resp, err := suite.client.Get("https://example.com")
	suite.Nil(err)
	suite.Equal(resp.StatusCode, http.StatusBadGateway)

	suite.status = http.StatusOK
	resp, err = suite.client.Get("https://example.com")
	suite.Nil(err)
	suite.Equal(resp.StatusCode, http.StatusBadGateway) // TLS response
}

func TestServer(t *testing.T) {
	suite.Run(t, &ServerTestSuite{})
}

func TestServerProxyChain(t *testing.T) {
	suite.Run(t, &ServerProxyChainTestSuite{})
}
