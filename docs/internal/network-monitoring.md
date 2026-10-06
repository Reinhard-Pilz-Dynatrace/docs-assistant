# Network monitoring Developer Guide

## Behavior and ownership

The network-monitoring behavior is owned by the `product/network_monitoring` package.

`ShouldSample` returns false when monitoring is disabled; otherwise it samples every `100 / sample_rate_percent`-th connection, with the step rounded down. `ShouldCaptureDNS` requires both `enabled` and `capture_dns_queries`. `ParseConfig` rejects a `sample_rate_percent` outside 1 to 100.

DNS query capture is on by default in the checked-in configuration (`capture_dns_queries: true`), so enabling network monitoring is sufficient to record DNS queries. The code logic itself was not changed; only the checked-in default value changed.

## Code and configuration locations

- `product/network_monitoring/network_monitor.go`
- `product/network_monitoring/config.yaml`
- `product/network_monitoring/network_monitor_test.go`

The checked-in defaults are `enabled: false`, `sample_rate_percent: 10`, and `capture_dns_queries: true`.

## Tests and verification

Run `go test ./product/network_monitoring` to verify configuration, sampling, and DNS capture behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults, including that `capture_dns_queries` is `true`.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
