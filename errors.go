package httransform

import "errors"

var (
	ErrProxyAuthorization = errors.New("Cannot authenticate proxy user")
)
