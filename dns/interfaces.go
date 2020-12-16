package dns

import "context"

// Interface is an interface for cachind DNS resolver.
type Interface interface {
	// Lookup returns resolved IPs for given hostname/ips. Important
	// property is that this list is shuffled on each function
	// execution.
	Lookup(context.Context, string) ([]string, error)
}
