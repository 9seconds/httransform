package auth

import (
	stderrors "errors"

	"github.com/PumpkinSeed/errors"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/layers"
)

var ErrLayerCannotAuth = errors.New("cannot authenticate")

type Layer struct {
	Authenticator Auth
	Require       bool
}

func (l *Layer) OnRequest(ctx *layers.LayerContext) error {
	user, err := l.Authenticator.Auth(ctx)

	switch {
	case l.Require && stderrors.Is(err, ErrNoAuth):
		return errors.Wrap(ErrLayerCannotAuth, err)
	case err == nil:
		ctx.Set("auth_user", user)
		return nil
	}

	return nil
}

func (l *Layer) OnResponse(ctx *layers.LayerContext, err error) {
	if err != nil && stderrors.Is(err, ErrLayerCannotAuth) {
		ctx.SetSimpleResponse(fasthttp.StatusProxyAuthRequired, err.Error())
	}
}
