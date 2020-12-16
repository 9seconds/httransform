package dns

import (
	"context"
	"net"
	"time"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/9seconds/httransform/v2/errors"
	"github.com/valyala/fastrand"
)

// DNS is a caching resolver.
type DNS struct {
	resolver net.Resolver
	cache    cache.Interface
}

// Lookup to conform Interface.
func (d *DNS) Lookup(ctx context.Context, hostname string) (hosts []string, err error) {
	if hh := d.cache.Get(hostname); hh != nil {
		hosts = hh.([]string)
	} else {
		hosts, err = d.doLookup(ctx, hostname)

		if err != nil {
			return nil, errors.Annotate(err, "cannot resolve dns", "dns_lookup", 0)
		}

		d.cache.Add(hostname, hosts)
	}

	return d.shuffle(hosts), nil
}

func (d *DNS) doLookup(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) == nil {
		return d.resolver.LookupHost(ctx, hostname)
	}

	addrs, err := d.resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, err // nolint: wrapcheck
	}

	hostnames := make([]string, len(addrs))

	for i := range addrs {
		hostnames[i] = addrs[i].IP.String()
	}

	return hostnames, nil
}

func (d *DNS) shuffle(in []string) []string {
	if len(in) < 2 { // nolint: gomnd
		return in
	}

	// See https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_%22inside-out%22_algorithm
	out := make([]string, len(in))

	for i := range in {
		j := int(fastrand.Uint32n(uint32(i + 1)))

		if j != i {
			out[i] = out[j]
		}

		out[j] = in[i]
	}

	return out
}

// New returns a new instance of DNS cache.
func New(cacheSize int, cacheTTL time.Duration, evictCallback cache.EvictCallback) Interface {
	return &DNS{
		cache: cache.New(cacheSize, cacheTTL, evictCallback),
	}
}
