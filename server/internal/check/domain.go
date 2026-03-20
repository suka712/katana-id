package check

import (
	"context"
	"errors"
	"net"
	"strings"
)

// PopularTLDs is the list of TLDs checked by default.
var PopularTLDs = []string{
	"com", "io", "dev", "co", "net", "org", "app", "ai", "xyz", "me",
}

// DomainChecker returns a CheckFn for a specific TLD.
// It uses DNS lookup — NXDOMAIN means the domain is likely available.
func DomainChecker(tld string) CheckFn {
	platform := DomainPlatform(tld)
	return func(ctx context.Context, name string) Result {
		domain := strings.ToLower(name) + "." + tld
		_, err := net.DefaultResolver.LookupHost(ctx, domain)
		if err != nil {
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
				return Result{Name: name, Platform: platform, Available: true}
			}
			return Result{Name: name, Platform: platform, Err: err.Error()}
		}
		return Result{Name: name, Platform: platform, Available: false}
	}
}

// DomainCheckers returns CheckFns for all PopularTLDs.
func DomainCheckers() []CheckFn {
	fns := make([]CheckFn, len(PopularTLDs))
	for i, tld := range PopularTLDs {
		fns[i] = DomainChecker(tld)
	}
	return fns
}
