package main

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"
)

type Resolver struct {
	resolver *net.Resolver
}

type NetworkString string

// Network type used for resolving hostnames
const (
	IP   NetworkString = "ip"
	IPv4 NetworkString = "ip4"
	IPv6 NetworkString = "ip6"
)

// Use an alternate dialer provided via `dnsServerAddr` string,
// specified without the port (53)
// instead of the default DNS server's address
func NewResolver(dnsServerAddr string) *Resolver {
	return &Resolver{
		resolver: &net.Resolver{
			PreferGo:     true, // 'false' seems to result in using the default (network's) DNS server, avoiding lookups via the IP address provided
			StrictErrors: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", dnsServerAddr+":53")
			},
		},
	}
}

// Resolves the `hostname` provided for the `network` (ip4|ip6|ip) provided and resolves the reverse
func (r *Resolver) ResolveHostname(ctx context.Context, network NetworkString, hostname string) {
	startTime := time.Now()

	ips, err := r.resolver.LookupIP(ctx, string(network), hostname)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok {
			LogError("Failed to resolve: %s: Error - '%s', was not found: %t\n", hostname, dnsErr.Err, dnsErr.IsNotFound)
		} else {
			LogError("Failed to resolve: %s Error - '%s'", hostname, err.Error())
		}
		return
	}

	LogInfo("IP addresses for hostname '%s': %v\n", hostname, addrString(ips))

	r.resolveReverse(ctx, ips, hostname)

	durationMs := time.Since(startTime).Milliseconds()
	LogInfo("Duration for resolving %s: %d ms\n", hostname, durationMs)
}

func (r *Resolver) ResolveHostnames(ctx context.Context, network NetworkString, hostnames []string) {
	var wg sync.WaitGroup
	for _, hostname := range hostnames {
		wg.Add(1)
		go func() {
			r.ResolveHostname(ctx, network, hostname)
			wg.Done()
		}()
	}
	wg.Wait()
}

// perform a reverse lookup for each ip address
func (r *Resolver) resolveReverse(ctx context.Context, ips []net.IP, hostname string) {
	blockedIpStr := "0.0.0.0"

	for _, ip := range ips {
		// ignore blocked hostnames
		if ip.Equal(net.ParseIP(blockedIpStr)) {
			if len(ips) == 1 {
				// we're done if this addr is the only IP addr.
				LogInfo("Ignoring attempt to resolve reverse for %s as it previously resolved to %s", hostname, blockedIpStr)
				return
			} else {
				// This is a remote possibility I suppose, but we'll handle it anyway in the rare event it occurs?
				continue
			}
		}

		names, err := r.resolver.LookupAddr(ctx, ip.String())
		if err != nil {
			if dnsErr, ok := err.(*net.DNSError); ok {
				LogError("Error performing reverse lookup for %s (%s): Error - '%s', was not found: %t\n", hostname, ip.String(), dnsErr.Err, dnsErr.IsNotFound)
			}
		} else {
			LogInfo("Reverse for %s (%s): %v", ip, hostname, strings.Join(names, ", "))
		}
	}
}

func addrString(ips []net.IP) string {
	addrStr := ""
	for i, ip := range ips {
		if i == len(ips)-1 {
			addrStr += ip.String() // avoid appending comma to last token
		} else {
			addrStr += ip.String() + ", "
		}
	}
	return addrStr
}
