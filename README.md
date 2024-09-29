Resolves hostnames (and performs reverse lookup) for n hostnames via the DNS server IP address provided.

`timeout` arg adds a timeout where attempts to resolve will be aborted if this duration is exceeded.

When `dnsserver` is not provided, the default resolver will be used.

`iptype` is an optional flag for resolving hostnames to `ipv4` `ipv6` or both. Defaults to `ip4`.

```bash
go build
./resolve-hostname [-dnsserver dns-server-ip-addr] [-timeout timeout-duration-ms] [-iptype ip|ip4|ip6] <hostname1> <hostname2> ...
```
