package auth

import "github.com/9seconds/httransform/v2/layers"

type noopAuth struct{}

func (n *noopAuth) Auth(_ *layers.LayerContext) (interface{}, error) {
	return nil, nil
}

func NewNoopAuth() Auth {
	return &noopAuth{}
}
