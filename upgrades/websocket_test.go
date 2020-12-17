package upgrades_test

import (
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/upgrades"
	"github.com/gobwas/ws/wsutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WebsocketTestSuite struct {
	UpgradesTestSuite
}

func (suite *WebsocketTestSuite) SetupTest() {
	suite.UpgradesTestSuite.SetupTest()

	suite.up = upgrades.AcquireWebsocket(upgrades.NoopWebsocketReactor{})
}

func (suite *WebsocketTestSuite) TearDownTest() {
	upgrades.ReleaseWebsocket(suite.up)

	suite.UpgradesTestSuite.TearDownTest()
}

func (suite *WebsocketTestSuite) TestPump() {
	wsutil.WriteClientText(&suite.clientConn.ReadBuffer, []byte("client"))
	suite.clientConn.On("Read", mock.Anything).Return(5, nil)
	suite.clientConn.On("Write", mock.Anything).Return(5, nil)
	suite.clientConn.On("Close").Return(nil)

	wsutil.WriteServerText(&suite.netlocConn.ReadBuffer, []byte("server"))
	suite.netlocConn.On("Read", mock.Anything).Return(5, nil)
	suite.netlocConn.On("Write", mock.Anything).Return(5, nil)
	suite.netlocConn.On("Close").Return(nil)

	suite.up.Manage(suite.clientConn, suite.netlocConn)
	time.Sleep(50 * time.Millisecond)

	msgs, err := wsutil.ReadClientMessage(&suite.netlocConn.WriteBuffer, nil)

	suite.NoError(err)
	suite.Len(msgs, 1)
	suite.Equal([]byte("client"), msgs[0].Payload)

	msgs, err = wsutil.ReadServerMessage(&suite.clientConn.WriteBuffer, nil)

	suite.NoError(err)
	suite.Len(msgs, 1)
	suite.Equal([]byte("server"), msgs[0].Payload)
}

func TestWebsocket(t *testing.T) {
	suite.Run(t, &WebsocketTestSuite{})
}
