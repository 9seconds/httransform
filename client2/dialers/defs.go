package dialers

import (
	"time"

	"github.com/rs/dnscache"

	"github.com/9seconds/httransform/v2/utils"
)

const dnsRefreshEvery = time.Minute

var (
	ErrBaseDialer            = &utils.Error{Text: "cannot dial"}
	ErrBaseDialerCannotParse = ErrBaseDialer.Extend("cannot parse address")
	ErrBaseDialerDNS         = ErrBaseDialer.Extend("dns lookup has failed")
	ErrBaseDialerNoIPs       = ErrBaseDialer.Extend("no ips are found")
	ErrBaseDialerAllFailed   = ErrBaseDialer.Extend("all ips are failed")

	ErrPooledDialerCanceled = ErrBaseDialer.Extend("context is closed")

	dns = func() *dnscache.Resolver {
		dns := &dnscache.Resolver{}

		go func() {
			for range time.Tick(dnsRefreshEvery) {
				dns.Refresh(true)
			}
		}()

		return dns
	}()
)
