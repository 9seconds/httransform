package auth

import (
	"time"

	"github.com/PumpkinSeed/errors"
)

const (
	AuthCacheFor            = time.Hour
	AuthCacheSizeMultiplier = 2
)

var (
	ErrAuth   = errors.New("cannot authenticate")
	ErrNoAuth = errors.New("no authentiation is set")

	ErrBasicAuthMalformed = errors.Wrap(errors.New("malformed header"), ErrAuth)
	ErrBasicAuthScheme    = errors.Wrap(errors.New("incorrect scheme"), ErrAuth)
	ErrBasicAuthPayload   = errors.Wrap(errors.New("incorrect payload"), ErrAuth)
	ErrBasicAuthDelimiter = errors.Wrap(errors.New("incorrect delimiter"), ErrAuth)
	ErrBasicAuthNoUser    = errors.Wrap(errors.New("no such user"), ErrAuth)

	ErrLayerCannotAuth = errors.New("cannot authenticate")
)
