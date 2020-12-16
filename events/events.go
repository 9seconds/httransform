package events

import (
	"fmt"
	"time"
)

// EventType is a unique identifier of the event type. There is a set of
// predefined events which are raised by httransform itself + users can
// define their own constants which are started from EventTypeUserBase.
type EventType byte

const (
	// EventTypeNotSet defines an empty event. If you see this type
	// somewhere, it is probably a bug.
	EventTypeNotSet EventType = iota

	// EventTypeCommonError defines a common errors produced by HTTP
	// server: cannot read request, timeouts on reading/writing, client
	// disconnects.
	//
	// Corresponding value is CommonErrorMeta instance.
	EventTypeCommonError

	// EventTypeNewCertificate defines an event when new TLS certificate
	// is GENERATED.
	//
	// Corresponding value is hostname (string).
	EventTypeNewCertificate

	// EventTypeDropCertificate defines an event when we evict TLS
	// certificate by either TTL or cache size limitation.
	//
	// Corresponding value is hostname (string).
	EventTypeDropCertificate

	// EventTypeFailedAuth is generated if user can't be authorized
	// by auth.Interface implementation.
	//
	// Corresponding value is nil (have no idea what to put there, tbh).
	EventTypeFailedAuth

	// EventTypeStartRequest is generated when auth is completed and
	// we just started to process a request.
	//
	// Corresponding value is RequestMeta instance.
	EventTypeStartRequest

	// EventTypeFailedRequest is generated when request is failed
	// for some logical reason (timeout etc).
	//
	// Corresponding value is ErrorMeta instance.
	EventTypeFailedRequest

	// EventTypeFinishRequest is generated when request is finished
	// OK and as expected.
	//
	// Corresponding value is ResponseMeta instance.
	EventTypeFinishRequest

	// EventTypeTraffic is generated when we've collected all traffic
	// for the request. Please pay attention that it could be that
	// this event will arrive after EventTypeFinishRequest.
	//
	// Corresponding value is TrafficMeta instance.
	EventTypeTraffic

	// EventTypeUserBase defines a constant you should use
	// to define your own event types.
	EventTypeUserBase
)

// IsUser returns if this event type is user one or predefined.
func (e EventType) IsUser() bool {
	return e >= EventTypeUserBase
}

// String conforms fmt.Stringer interface.
func (e EventType) String() string {
	switch e {
	case EventTypeNotSet:
		return "NOT_SET"
	case EventTypeCommonError:
		return "COMMON_ERROR"
	case EventTypeNewCertificate:
		return "NEW_CERTIFICATE"
	case EventTypeDropCertificate:
		return "DROP_CERTIFICATE"
	case EventTypeFailedAuth:
		return "FAILED_AUTH"
	case EventTypeStartRequest:
		return "START_REQUEST"
	case EventTypeFailedRequest:
		return "FAILED_REQUEST"
	case EventTypeFinishRequest:
		return "FINISH_REQUEST"
	case EventTypeTraffic:
		return "TRAFFIC"
	case EventTypeUserBase:
	}

	return fmt.Sprintf("USER(%d)", e-EventTypeUserBase)
}

// Event defines event information.
type Event struct {
	// Type defines a type of the event.
	Type EventType

	// Time defines a time when this event was generated.
	Time time.Time

	// Value defines an attached value with additional information.
	Value interface{}
}

// String conforms fmt.Stringer interface.
func (e *Event) String() string {
	return fmt.Sprintf("%v %v -> %v", e.Type, e.Time, e.Value)
}

// IsUser is a shortcut for evt.Type.IsUser.
func (e *Event) IsUser() bool {
	return e.Type.IsUser()
}
