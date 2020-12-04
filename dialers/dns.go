package dialers

import (
	"context"
	"fmt"
	"net"

	"github.com/9seconds/httransform/v2/cache"
	"github.com/valyala/fastrand"
)

type dnsCache struct {
	resolver net.Resolver
	cache    cache.Interface
}

func (d *dnsCache) Lookup(ctx context.Context, hostname string) (hosts []string, err error) {
	if hh := d.cache.Get(hostname); hh != nil {
		hosts = hh.([]string)
	} else {
		hosts, err = d.doLookup(ctx, hostname)

		if err != nil {
			return nil, fmt.Errorf("cannot resolve dns for hostname: %w", err)
		}

		d.cache.Add(hostname, hosts)
	}

	return d.shuffle(hosts), nil
}

func (d *dnsCache) doLookup(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) == nil {
		return d.resolver.LookupHost(ctx, hostname)
	}

	addrs, err := d.resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, err
	}

	hostnames := make([]string, len(addrs))

	for i := range addrs {
		hostnames[i] = addrs[i].IP.String()
	}

	return hostnames, nil
}

func (d *dnsCache) shuffle(in []string) []string {
	if len(in) < 2 {
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
