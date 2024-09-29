Resolves hostnames (and performs reverse lookup) for n hostnames via the DNS server IP address provided.

`timeout` arg adds a timeout where attempts to resolve will be aborted if this duration is exceeded.

When `dnsserver` is not provided, the default resolver will be used.

```bash
go build
./resolve-hostname [-dnsserver dns-server-ip-addr] [-timeout timeout-duration-ms] <hostname1> <hostname2> ...
```
