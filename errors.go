package httransform

import "golang.org/x/xerrors"

var (
	// ErrProxyAuthorization is the error for ProxyAuthorizationBasicLayer
	// instance. If OnRequest callback of this method returns such an error,
	// then OnResponse callback generates correct 407 response.
	ErrProxyAuthorization = xerrors.New("cannot authenticate proxy user")
)
