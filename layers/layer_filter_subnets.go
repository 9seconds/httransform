package layers

import (
	"fmt"
	"net"

	"github.com/9seconds/httransform/v2/dns"
	"github.com/9seconds/httransform/v2/errors"
	"github.com/kentik/patricia"
	"github.com/kentik/patricia/bool_tree"
	"github.com/valyala/fasthttp"
)

type filterSubnetsLayer struct {
	v4 *bool_tree.TreeV4
	v6 *bool_tree.TreeV6
}

func (f *filterSubnetsLayer) OnRequest(ctx *Context) error {
	host, _, _ := net.SplitHostPort(ctx.ConnectTo)

	resolved, err := dns.Default.Lookup(ctx, host)
	if err != nil {
		// pass unresolved name, delegate it to executor.
		return nil // nolint: nilerr
	}

	for _, v := range resolved {
		if f.filterIP(net.ParseIP(v)) {
			return &errors.Error{
				StatusCode: fasthttp.StatusForbidden,
				Message:    fmt.Sprintf("host %s has filtered IP %s", host, v),
				Code:       "filtered_subnet",
			}
		}
	}

	return nil
}

func (f *filterSubnetsLayer) filterIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return f.filterIPv4(ip4)
	}

	return f.filterIPv6(ip)
}

func (f *filterSubnetsLayer) filterIPv4(addr net.IP) bool {
	ip := patricia.NewIPv4AddressFromBytes(addr, 32) // nolint: gomnd

	if ok, _, err := f.v4.FindDeepestTag(ip); ok && err == nil {
		return true
	}

	return false
}

func (f *filterSubnetsLayer) filterIPv6(addr net.IP) bool {
	ip := patricia.NewIPv6Address(addr, 128) // nolint: gomnd

	if ok, _, err := f.v6.FindDeepestTag(ip); ok && err == nil {
		return true
	}

	return false
}

func (f *filterSubnetsLayer) OnResponse(ctx *Context, err error) error {
	return err
}

// NewFilterSubnetsLayer filters out requests which should not be passed
// further. You can use it to protect an access to your private networks.
//
// For example, you can block requests to 127.0.0.1/8, 10.0.0.0/8.
//
// This layer does DNS queries and uses their results to understand if
// it worth to proceed or not.
func NewFilterSubnetsLayer(subnets []net.IPNet) (Layer, error) {
	instance := &filterSubnetsLayer{
		v4: bool_tree.NewTreeV4(),
		v6: bool_tree.NewTreeV6(),
	}

	for i := range subnets {
		v4, v6, err := patricia.ParseFromIPAddr(&subnets[i])

		switch {
		case err != nil:
			return nil, fmt.Errorf("incorrect subnet %v: %w", subnets[i], err)
		case v4 != nil:
			if _, _, err := instance.v4.Set(*v4, true); err != nil {
				return nil, fmt.Errorf("cannot set v4 subnet %v: %w", subnets[i], err)
			}
		case v6 != nil:
			if _, _, err := instance.v6.Set(*v6, true); err != nil {
				return nil, fmt.Errorf("cannot set v6 subnet %v: %w", subnets[i], err)
			}
		}
	}

	return instance, nil
}
