package dialers

import (
	"time"
)

const (
	// DefaultTimeout defines a timeout which should be used to
	// establish a TCP connections to target netlocs if user provides no
	// value.
	DefaultTimeout = 20 * time.Second

	// DefaultDNSUpdateEvery defines a time period which we use to refresh
	// DNS entries in a local cache if user provides no value.
	DefaultDNSUpdateEvery = time.Minute
)

// Opts define a set of common options for each dialer. This struct here
// is just to avoid a huge method signature and boilerplate code about
// default values for each and every dialer.
type Opts struct {
	// Timeout defines a timeout which use to establish TCP connections to
	// target netlocs. It is also a timeout on UpgradeToTLS operations.
	Timeout time.Duration

	// DNSUpdateEvery defines a time period which we use to refresh DNS
	// entries in a local cache.
	DNSUpdateEvery time.Duration

	// TLSSkipVerify defines if we want to skip verification of
	// TLScertificates or not.
	//
	// People like to disable it so I just wanna make a yet another general
	// announcement that it can be a bad practice from securtity POV.
	//
	// You was warned.
	TLSSkipVerify bool
}

// GetTimeout returns a timeout value or fallbacks to default one.
func (o *Opts) GetTimeout() time.Duration {
	if o.Timeout == 0 {
		return DefaultTimeout
	}

	return o.Timeout
}

// GetDNSUpdateEvery returns a value for DNS cache refresh or fallbacks
// to default one.
func (o *Opts) GetDNSUpdateEvery() time.Duration {
	if o.DNSUpdateEvery == 0 {
		return DefaultDNSUpdateEvery
	}

	return o.DNSUpdateEvery
}

// GetTLSSkipVerify return a value for skipping TLS verification or
// fallbacks to default one.
func (o *Opts) GetTLSSkipVerify() bool {
	return o.TLSSkipVerify
}
