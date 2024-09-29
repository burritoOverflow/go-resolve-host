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
	"time"
)

const helpMsg string = `Resolve hostnames via a provided DNS address; cancel if not complete by timeout:
Usage: resolve-hostname [-dnsserver dns-server-ip-addr] [-timeout timeout-duration-ms] <hostname1> <hostname2> ...`

// ensure this is a valid ip address
// we have a valid IP provided for DNS; create our resolver for this
// otherwise, we'll use the default DNS server
func getDnsResolver(dnsServerIp *string) (*Resolver, error) {
	// use our `Resolver` if addr present and valid
	if dnsServerIp != nil && len(*dnsServerIp) != 0 {
		if !(net.ParseIP(*dnsServerIp) != nil) {
			return nil, errors.New(fmt.Sprintf("Invalid ip address: %s", *dnsServerIp))
		} else {
			return NewResolver(*dnsServerIp), nil
		}
	}

	// otherwise, use the default
	return &Resolver{
		resolver: net.DefaultResolver,
	}, nil
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

func validNetworkString(s string) bool {
	switch NetworkString(s) {
	case IP, IPv4, IPv6:
		return true
	default:
		return false
	}
}

func main() {
	InitializeLogger()
	totalStart := time.Now()

	// this is a bit short by default
	defaultTimeoutMs := 1000

	dnsServerIp := flag.String("dnsserver", "", "The DNS server to use to resolve hostnames")
	timeoutArg := flag.Int("timeout", defaultTimeoutMs, "Timeout in milliseconds")
	networkType := flag.String("iptype", string(IPv4), "Resolve ipv4, ipv6, or both. Must be one 'ip', 'ip4', or 'ip6' (default 'ip4')")
	flag.Parse()

	if *timeoutArg < 0 {
		LogError("Invalid value provided for timeout: '%d'\n", *timeoutArg)
		log.Fatalf(helpMsg)
	}

	if !validNetworkString(*networkType) {
		LogError("Invalid value provided for network string: '%s'\n", *networkType)
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

	r.ResolveHostnames(ctx, NetworkString(*networkType), hostnames)

	totalDuration := time.Since(totalStart)
	addrs := strings.Join(hostnames, ", ")

	LogInfo("%s for %d addresses (%s): %d ms\n", prefixStr(totalDuration, timeout), len(hostnames), addrs, totalDuration.Milliseconds())
}
