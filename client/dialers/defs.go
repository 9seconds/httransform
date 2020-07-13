package dialers

import "github.com/PumpkinSeed/errors"

var (
	ErrDialer = errors.New("cannot dial")
	ErrCannotConnectToProxy = errors.Wrap(errors.New("cannot connect to proxy"), ErrDialer)
	ErrCannotReadResponse = errors.Wrap(errors.New("cannot read proxy response"), ErrDialer)
	ErrProxyRejected = errors.Wrap(errors.New("proxy rejected connect request"), ErrDialer)
)
