package auth

import (
	"fmt"
	"net"

	"github.com/kentik/patricia"
	"github.com/kentik/patricia/string_tree"
	"github.com/valyala/fasthttp"
)

type ipWhitelist struct {
	v4 *string_tree.TreeV4
	v6 *string_tree.TreeV6
}

func (i *ipWhitelist) Authenticate(ctx *fasthttp.RequestCtx) (string, error) {
	ip := ctx.RemoteIP()
	ip4 := ip.To4()

	if ip4 != nil {
		addr := patricia.NewIPv4AddressFromBytes(ip4, 32)

		if ok, user, err := i.v4.FindDeepestTag(addr); ok && err == nil {
			return user, nil
		}

		return "", ErrFailedAuth
	}

	addr := patricia.NewIPv6Address(ip, 128)

	if ok, user, err := i.v6.FindDeepestTag(addr); ok && err == nil {
		return user, nil
	}

	return "", ErrFailedAuth
}

// NewIPWhitelist returns an implementation of authenicator which does
// auth based on a user IP address.
//
// An input parameter is a map where key is the name of the user and
// values - an array of subnets which are associated with that user. So,
// if incoming request is established from that subnet, we associate it
// with a user.
//
// This authenticator is implemented to work with RequestCtx with no
// normalization.
func NewIPWhitelist(tags map[string][]net.IPNet) (Interface, error) {
	instance := &ipWhitelist{
		v4: string_tree.NewTreeV4(),
		v6: string_tree.NewTreeV6(),
	}

	for user, whitelists := range tags {
		for i := range whitelists {
			v4, v6, err := patricia.ParseFromIPAddr(&whitelists[i])

			switch {
			case err != nil:
				return nil, fmt.Errorf("incorrect subnet: %w", err)
			case v4 != nil:
				if _, _, err := instance.v4.Set(*v4, user); err != nil {
					return nil, fmt.Errorf("cannot set v4 subnet %v: %w", whitelists[i], err)
				}
			case v6 != nil:
				if _, _, err := instance.v6.Set(*v6, user); err != nil {
					return nil, fmt.Errorf("cannot set v6 subnet %v: %w", whitelists[i], err)
				}
			}
		}
	}

	return instance, nil
}
