package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const helpMsg string = `Resolve hostnames via a provided DNS address; cancel if not complete by timeout:
Usage: resolve-hostname [-dnsserver dns-server-ip-addr] [-timeout timeout-duration-ms] <hostname1> <hostname2> ...`

type Resolver struct {
	resolver *net.Resolver
}

// perform a reverse lookup for each ip address
func (r *Resolver) resolveReverse(ctx context.Context, ips []net.IP, hostname string) {
	for _, ip := range ips {
		// ignore blocked hostnames
		blockIpStr := "0.0.0.0"
		if ip.Equal(net.ParseIP(blockIpStr)) {
			LogInfo("Ignoring attempt to resolve reverse for %s as it previously resolved to %s", hostname, blockIpStr)
			continue
		}

		names, err := r.resolver.LookupAddr(ctx, ip.String())
		if err != nil {
			LogError("Error performing reverse lookup for %s (%s): %v", ip, hostname, err)
		} else {
			LogInfo("Reverse for %s (%s): %v", ip, hostname, strings.Join(names, ", "))
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

func resolveHostnames(ctx context.Context, hostnames []string, r *Resolver) {
	var wg sync.WaitGroup
	for _, hostname := range hostnames {
		wg.Add(1)
		go func() {
			r.resolveHostname(ctx, hostname)
			wg.Done()
		}()
	}
	wg.Wait()
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
				d := net.Dialer{}
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

// ensure this is a valid ip address
// we have a valid IP provided for DNS; create our resolver for this
// otherwise, we'll use the default DNS server
func getDnsResolver(dnsServerIp *string) (*Resolver, error) {
	r := Resolver{}

	if len(*dnsServerIp) != 0 {
		if !(net.ParseIP(*dnsServerIp) != nil) {
			return nil, errors.New(fmt.Sprintf("Invalid ip address: %s", *dnsServerIp))
		} else {
			r = newResolver(*dnsServerIp)
		}
	} else {
		r.resolver = net.DefaultResolver
	}

	return &r, nil
}

func prefixStr(total time.Duration, timeout time.Duration) string {
	prefixStr := ""
	if total > timeout {
		prefixStr = "Deadline exceeded"
	} else {
		prefixStr = "Total duration"
	}
	return prefixStr
}

func main() {
	InitializeLogger()
	totalStart := time.Now()

	// this is a bit short by default
	defaultTimeoutMs := 1000

	dnsServerIp := flag.String("dnsserver", "", "The DNS server to use to resolve hostnames")
	timeoutArg := flag.Int("timeout", defaultTimeoutMs, "Timeout in milliseconds")
	flag.Parse()

	if *timeoutArg < 0 {
		LogError("Invalid value provided for timeout: %d\n", *timeoutArg)
		log.Fatalf(helpMsg)
	}

	// only hostnames are required
	hostnames := flag.Args()
	if len(hostnames) == 0 {
		log.Fatalf(helpMsg)
	}

	r, err := getDnsResolver(dnsServerIp)
	if err != nil {
		LogError(err.Error())
		os.Exit(1)
	}

	timeout := time.Duration(*timeoutArg) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resolveHostnames(ctx, hostnames, r)

	totalDuration := time.Since(totalStart)
	addrs := strings.Join(hostnames, ", ")

	LogInfo("%s for %d addresses %d ms: (%s)\n", prefixStr(totalDuration, timeout), len(hostnames), totalDuration.Milliseconds(), addrs)
}
