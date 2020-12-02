package events_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/events"
	"github.com/stretchr/testify/suite"
)

type NoopProcessorFactoryTestSuite struct {
	suite.Suite
}

func (suite *NoopProcessorFactoryTestSuite) TestOk() {
	proc := events.NoopProcessorFactory()
	proc.Process(nil)
	proc.Shutdown()
}

func TestNoopProcessorFactory(t *testing.T) {
	suite.Run(t, &NoopProcessorFactoryTestSuite{})
}
