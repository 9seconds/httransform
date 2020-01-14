package dialers

import "github.com/9seconds/httransform/v2/utils"

var (
	ErrBaseDialer            = &utils.Error{Text: "cannot dial"}
	ErrBaseDialerCannotParse = ErrBaseDialer.Extend("cannot parse address")
	ErrBaseDialerDNS         = ErrBaseDialer.Extend("dns lookup has failed")
	ErrBaseDialerNoIPs       = ErrBaseDialer.Extend("no ips are found")
	ErrBaseDialerAllFailed   = ErrBaseDialer.Extend("all ips are failed")

	ErrPooledDialerCanceled = ErrBaseDialer.Extend("context is closed")
)
