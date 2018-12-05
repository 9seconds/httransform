package httransform

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NoopLoggerTestSuite struct {
	suite.Suite

	logger *NoopLogger
}

func (suite *NoopLoggerTestSuite) SetupTest() {
	suite.logger = &NoopLogger{}
}

func (suite *NoopLoggerTestSuite) TestInterface() {
	suite.Implements((*Logger)(nil), suite.logger)
}

func (suite *NoopLoggerTestSuite) TestDummy() {
	suite.logger.Debug("msg")
	suite.logger.Info("msg")
	suite.logger.Warn("msg")
	suite.logger.Error("msg")
	suite.Panics(func() { suite.logger.Panic("msg") })
}

type StdLoggerTestSuite struct {
	suite.Suite

	logger *StdLogger
	output *bytes.Buffer
}

func (suite *StdLoggerTestSuite) SetupTest() {
	suite.output = &bytes.Buffer{}
	suite.logger = &StdLogger{
		Log: log.New(suite.output, "", 0),
	}
}

func (suite *StdLoggerTestSuite) TestInterface() {
	suite.Implements((*Logger)(nil), suite.logger)
}

func (suite *StdLoggerTestSuite) TestDebug() {
	suite.logger.Debug("Msg: %s", "1")

	suite.Equal([]byte("Msg: 1\n"), suite.output.Bytes())
}

func (suite *StdLoggerTestSuite) TestInfo() {
	suite.logger.Info("Msg: %s", "1")

	suite.Equal([]byte("Msg: 1\n"), suite.output.Bytes())
}

func (suite *StdLoggerTestSuite) TestWarn() {
	suite.logger.Warn("Msg: %s", "1")

	suite.Equal([]byte("Msg: 1\n"), suite.output.Bytes())
}

func (suite *StdLoggerTestSuite) TestError() {
	suite.logger.Error("Msg: %s", "1")

	suite.Equal([]byte("Msg: 1\n"), suite.output.Bytes())
}

func (suite *StdLoggerTestSuite) TestPanic() {
	suite.Panics(func() { suite.logger.Panic("Msg: %s", "1") })

	suite.Equal([]byte("Msg: 1\n"), suite.output.Bytes())
}

func TestNoopLogger(t *testing.T) {
	suite.Run(t, &NoopLoggerTestSuite{})
}

func TestStdLogger(t *testing.T) {
	suite.Run(t, &StdLoggerTestSuite{})
}
