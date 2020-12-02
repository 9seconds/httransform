package events_test

import (
	"testing"

	"github.com/9seconds/httransform/v2/events"
	"github.com/stretchr/testify/suite"
)

type EventTypeTestSuite struct {
	suite.Suite
}

func (suite *EventTypeTestSuite) TestIsUser() {
	suite.False(events.EventTypeNotSet.IsUser())
	suite.False(events.EventTypeCommonError.IsUser())
	suite.False(events.EventTypeNewCertificate.IsUser())
	suite.False(events.EventTypeDropCertificate.IsUser())
	suite.False(events.EventTypeFailedAuth.IsUser())
	suite.False(events.EventTypeStartRequest.IsUser())
	suite.False(events.EventTypeFailedRequest.IsUser())
	suite.False(events.EventTypeFinishRequest.IsUser())
	suite.False(events.EventTypeTraffic.IsUser())

	suite.True(events.EventTypeUserBase.IsUser())
	suite.True((events.EventTypeUserBase + 1).IsUser())
}

func (suite *EventTypeTestSuite) TestString() {
	suite.Equal("NOT_SET", events.EventTypeNotSet.String())
	suite.Equal("COMMON_ERROR", events.EventTypeCommonError.String())
	suite.Equal("NEW_CERTIFICATE", events.EventTypeNewCertificate.String())
	suite.Equal("DROP_CERTIFICATE", events.EventTypeDropCertificate.String())
	suite.Equal("FAILED_AUTH", events.EventTypeFailedAuth.String())
	suite.Equal("START_REQUEST", events.EventTypeStartRequest.String())
	suite.Equal("FAILED_REQUEST", events.EventTypeFailedRequest.String())
	suite.Equal("FINISH_REQUEST", events.EventTypeFinishRequest.String())
	suite.Equal("TRAFFIC", events.EventTypeTraffic.String())

	suite.Equal("USER(0)", events.EventTypeUserBase.String())
	suite.Equal("USER(1)", (1 + events.EventTypeUserBase).String())
}

func TestEventType(t *testing.T) {
	suite.Run(t, &EventTypeTestSuite{})
}
