# Heartbeat Developer Guide

## Behavior and ownership

The heartbeat behavior is owned by the `product/heartbeat` package.

`IsHeartbeatDue` returns true when the heartbeat is enabled and at least `heartbeat_interval_seconds` have passed since the last one. `ShouldIncludeHostMetadata` requires both `enabled` and `include_host_metadata`. `ParseConfig` rejects a `heartbeat_interval_seconds` value outside 5 to 600.

## Code and configuration locations

- `product/heartbeat/heartbeat.go`
- `product/heartbeat/config.yaml`
- `product/heartbeat/heartbeat_test.go`

The checked-in defaults are `enabled: true`, `heartbeat_interval_seconds: 20`, and `include_host_metadata: true`.

## Tests and verification

Run `go test ./product/heartbeat` to verify configuration and behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
