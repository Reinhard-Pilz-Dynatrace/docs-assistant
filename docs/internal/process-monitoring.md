# Process monitoring Developer Guide

## Behavior and ownership

The process-monitoring behavior is owned by the `product/process_monitoring` package.

The checked-in default `process_monitoring.scan_interval_seconds` is 60 (previously 30). `process_monitoring.enabled` remains `true` by default. `ParseConfig` rejects a `scan_interval_seconds` that is zero or less. `include_container_processes` defaults to `false` when absent, so container processes are excluded unless configured.

## Code and configuration locations

- `product/process_monitoring/process_monitor.go`
- `product/process_monitoring/config.yaml`
- `product/process_monitoring/process_monitor_test.go`

The default scan interval is 60 seconds, set in `product/process_monitoring/config.yaml`.

## Tests and verification

Run `go test ./product/process_monitoring` to verify configuration and process selection behavior.

`TestCheckedInConfigUsesSixtySecondInterval` reads the checked-in `config.yaml` and asserts that the scan interval is 60. `TestParseConfigReadsCheckedInDefault` parses an inline YAML string with a 30-second value, not the checked-in file, so it does not verify the checked-in default.

## Operational notes

This change affects the default cadence only. No measured CPU or performance improvement is claimed. Operational diagnostics and platform-specific behavior are not established yet.
