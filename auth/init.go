package auth

import "github.com/PumpkinSeed/errors"

var ErrAuth = errors.New("cannot authenticate")

type Auth interface {
	Auth(string) (interface{}, error)
}
