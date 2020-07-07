package dialers

import (
	"time"

	"github.com/rs/dnscache"
)

var (
	dns = func() *dnscache.Resolver {
		dns := &dnscache.Resolver{}

		go func() {
			for range time.Tick(DNSRefreshEvery) {
				dns.Refresh(true)
			}
		}()

		return dns
	}()
)
