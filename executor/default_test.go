package executor_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/gorilla/websocket"
	"github.com/mccutchen/go-httpbin/httpbin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type EventChannelMock struct {
	mock.Mock
}

func (e *EventChannelMock) Send(ctx context.Context, eventType events.EventType, value interface{}, shardKey string) {
	e.Called(ctx, eventType, value, shardKey)
}

type upgrader struct {
	websocket.Upgrader
}

func (u *upgrader) Handler(w http.ResponseWriter, r *http.Request) {
	c, err := u.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	c.WriteJSON(map[string]string{"hello": "world"})
}

type MakeDefaultExecutorTestSuite struct {
	suite.Suite

	endpoint      *httptest.Server
	fhttpCtx      *fasthttp.RequestCtx
	eventsChannel *EventChannelMock
	ctx           *layers.Context
	exec          executor.Executor
}

func (suite *MakeDefaultExecutorTestSuite) SetupSuite() {
	httpbinApp := httpbin.NewHTTPBin()
	mux := http.NewServeMux()
	wsUpgrader := &upgrader{}

	mux.HandleFunc("/ip", httpbinApp.IP)
	mux.HandleFunc("/ws", wsUpgrader.Handler)

	suite.endpoint = httptest.NewServer(mux)
}

func (suite *MakeDefaultExecutorTestSuite) TearDownSuite() {
	suite.endpoint.Close()
}

func (suite *MakeDefaultExecutorTestSuite) SetupTest() {
	suite.fhttpCtx = &fasthttp.RequestCtx{}

	remoteAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 65342,
	}

	suite.fhttpCtx.Init(&fasthttp.Request{}, remoteAddr, nil)

	suite.ctx = layers.AcquireContext()
	suite.eventsChannel = &EventChannelMock{}

	suite.ctx.Init(suite.fhttpCtx,
		"127.0.0.1:8000",
		suite.eventsChannel,
		"user",
		events.RequestTypeTLS)

	suite.ctx.ConnectTo = suite.endpoint.Listener.Addr().String()

	suite.exec = executor.MakeDefaultExecutor(dialers.NewBase(dialers.Opts{}))
}

func (suite *MakeDefaultExecutorTestSuite) TearDownTest() {
	suite.eventsChannel.AssertExpectations(suite.T())
	layers.ReleaseContext(suite.ctx)
}

func (suite *MakeDefaultExecutorTestSuite) TestIncorrectAddress() {
	suite.ctx.ConnectTo = ""

	suite.Error(suite.exec(suite.ctx))
	suite.Len(suite.ctx.Response().Body(), 0)
}

func (suite *MakeDefaultExecutorTestSuite) TestCannotDial() {
	suite.ctx.ConnectTo += "1"

	suite.Error(suite.exec(suite.ctx))
	suite.Len(suite.ctx.Response().Body(), 0)
}

func (suite *MakeDefaultExecutorTestSuite) TestSendHTTP() {
	suite.eventsChannel.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	req := suite.ctx.Request()

	req.SetRequestURI(suite.endpoint.URL + "/ip")
	req.Header.SetHost("127.0.0.1")

	suite.NoError(suite.exec(suite.ctx))

	v := map[string]interface{}{}

	suite.NoError(json.Unmarshal(suite.ctx.Response().Body(), &v))
}

func (suite *MakeDefaultExecutorTestSuite) TestUpgrade() {
	suite.eventsChannel.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	req := suite.ctx.Request()

	req.SetRequestURI(suite.endpoint.URL + "/ws")
	req.Header.SetHost("127.0.0.1")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-Websocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "aaa")

	suite.NoError(suite.exec(suite.ctx))
}

func TestMakeDefaultExecutor(t *testing.T) {
	suite.Run(t, &MakeDefaultExecutorTestSuite{})
}
