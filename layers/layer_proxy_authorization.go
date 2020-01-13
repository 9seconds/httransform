package layers

import (
	"crypto/subtle"
	"errors"

	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/utils"
)

type ProxyAuthorizationLayer struct {
	User     string
	Password string
	Realm    string
}

func (p *ProxyAuthorizationLayer) OnRequest(ctx *LayerContext) error {
	value, ok := ctx.RequestHeaders.GetString("Proxy-Authorization")
	if !ok {
		return ErrProxyAuthorizationCannotGetHeader
	}

	user, password, err := utils.ExtractAuthentication(value)
	if err != nil {
		return ErrProxyAuthorizationCannotExtract
	}

	userRes := subtle.ConstantTimeCompare([]byte(p.User), []byte(user)) != 1
	passRes := subtle.ConstantTimeCompare([]byte(p.Password), []byte(password)) != 1

	if userRes || passRes {
		return ErrProxyAuthorizationIncorrect
	}

	return nil
}

func (p *ProxyAuthorizationLayer) OnResponse(ctx *LayerContext, err error) {
	if errors.Is(err, ErrProxyAuthorization) {
		ctx.Response.ResetBody()
		ctx.Response.SetStatusCode(fasthttp.StatusProxyAuthRequired)

		if p.Realm != "" {
			ctx.ResponseHeaders.SetString("Proxy-Authenticate", "Basic Realm=\""+p.Realm+"\"")
		} else {
			ctx.ResponseHeaders.SetString("Proxy-Authenticate", "Basic")
		}
	}
}
