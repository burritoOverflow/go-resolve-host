package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const helpMsg string = `Resolve hostnames via a provided DNS address:
Usage: resolve-hostname <dns-server-ip-addr> <hostname1> <hostname2> ...`

type Resolver struct {
	resolver *net.Resolver
}

// perform a reverse lookup for each ip address
func (r *Resolver) resolveReverse(ctx context.Context, ips []net.IP, hostname string) {
	for _, ip := range ips {
		names, err := r.resolver.LookupAddr(ctx, ip.String())
		if err != nil {
			LogError("Error performing reverse lookup for %s (%s): %v", ip, hostname, err)
		} else {
			LogError("Reverse for %s (%s): %v", ip, hostname, strings.Join(names, ", "))
		}
	}
}

func (r *Resolver) resolveHostname(ctx context.Context, hostname string) {
	startTime := time.Now()

	ips, err := r.resolver.LookupIP(ctx, "ip4", hostname)
	if err != nil {
		LogError("Failed to resolve %s: %v\n", hostname, err)
		return
	}

	LogInfo("IP addresses for %s: %v\n", hostname, addrString(ips))

	r.resolveReverse(ctx, ips, hostname)

	durationMs := time.Since(startTime).Milliseconds()
	LogInfo("Duration for resolving %s: %d ms\n", hostname, durationMs)
}

// Use an alternate dialer provided via `dnsServerAddr` string,
// specified without the port (53)
// instead of the default DNS server's address
func newResolver(dnsServerAddr string) Resolver {
	return Resolver{
		resolver: &net.Resolver{
			PreferGo:     true, // 'false' seems to result in using the default (network's) DNS server, avoiding lookups via the IP address provided
			StrictErrors: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: time.Second * 5}
				return d.DialContext(ctx, "udp", dnsServerAddr+":53")
			},
		},
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

func main() {
	InitializeLogger()
	totalStart := time.Now()

	if len(os.Args) < 3 {
		fmt.Println(helpMsg)
		return
	}

	// ensure this is a valid ip address
	dnsServerIp := os.Args[1]
	if !(net.ParseIP(dnsServerIp) != nil) {
		LogError("Invalid ip address: %s", dnsServerIp)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	r := newResolver(dnsServerIp)

	hostnames := os.Args[2:]

	for _, hostname := range hostnames {
		wg.Add(1)
		go func() {
			r.resolveHostname(ctx, hostname)
			wg.Done()
		}()
	}

	wg.Wait()

	totalDuration := time.Since(totalStart)
	addrs := strings.Join(os.Args[2:], ", ")
	LogInfo("Total duration for %d addresses: %s: %d ms\n", len(os.Args[2:]), addrs, totalDuration.Milliseconds())
}
