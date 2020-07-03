package auth

import "github.com/9seconds/httransform/v2/layers"

type noopAuth struct{}

func (n *noopAuth) Auth(_ *layers.LayerContext) (bool, interface{}, error) {
	return false, nil, nil
}

func NewNoopAuth() Auth {
	return &noopAuth{}
}
