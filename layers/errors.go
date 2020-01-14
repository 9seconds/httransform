package layers

import "github.com/9seconds/httransform/v2/utils"

var (
	ErrProxyAuthorization                = &utils.Error{Text: "cannot authorize this request"}
	ErrProxyAuthorizationCannotGetHeader = ErrProxyAuthorization.Extend("cannot get authorization header")
	ErrProxyAuthorizationCannotExtract   = ErrProxyAuthorization.Extend("cannot extract user/password")
	ErrProxyAuthorizationIncorrect       = ErrProxyAuthorization.Extend("incorrect authorization")
)
