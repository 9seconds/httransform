package auth

import "github.com/9seconds/httransform/v2/layers"

type Auth interface {
	Auth(*layers.LayerContext) (interface{}, error)
}
