package auth

import (
	"github.com/PumpkinSeed/errors"

	"github.com/9seconds/httransform/v2/layers"
)

var (
	ErrAuth   = errors.New("cannot authenticate")
	ErrNoAuth = errors.New("no authentiation is set")
)

type Auth interface {
	Auth(*layers.LayerContext) (interface{}, error)
}
