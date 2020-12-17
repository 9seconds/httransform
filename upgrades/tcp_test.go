package upgrades_test

import (
	"testing"
	"time"

	"github.com/9seconds/httransform/v2/upgrades"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TCPTestSuite struct {
	UpgradesTestSuite
}

func (suite *TCPTestSuite) SetupTest() {
	suite.UpgradesTestSuite.SetupTest()

	suite.up = upgrades.NewTCP(upgrades.NoopTCPReactor{})
}

func (suite *TCPTestSuite) TearDownTest() {
	upgrades.ReleaseTCP(suite.up)

	suite.UpgradesTestSuite.TearDownTest()
}

func (suite *TCPTestSuite) TestPump() {
	suite.clientConn.ReadBuffer.Write([]byte{1, 2, 3, 4, 5})
	suite.clientConn.On("Read", mock.Anything).Return(5, nil)
	suite.clientConn.On("Write", mock.Anything).Return(5, nil)
	suite.clientConn.On("Close").Return(nil)

	suite.netlocConn.ReadBuffer.Write([]byte{11, 12, 13, 14, 15})
	suite.netlocConn.On("Read", mock.Anything).Return(5, nil)
	suite.netlocConn.On("Write", mock.Anything).Return(5, nil)
	suite.netlocConn.On("Close").Return(nil)

	suite.up.Manage(suite.clientConn, suite.netlocConn)
	time.Sleep(50 * time.Millisecond)

	suite.Equal([]byte{1, 2, 3, 4, 5}, suite.netlocConn.WriteBuffer.Bytes())
	suite.Equal([]byte{11, 12, 13, 14, 15}, suite.clientConn.WriteBuffer.Bytes())
}

func TestTCP(t *testing.T) {
	suite.Run(t, &TCPTestSuite{})
}
