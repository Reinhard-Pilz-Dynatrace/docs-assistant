# Process monitoring Developer Guide

## Behavior and ownership

The process-monitoring behavior is owned by the `product/process_monitoring` package.

The checked-in default `process_monitoring.scan_interval_seconds` is 60 seconds (changed from 30 in PR #4). `enabled` remains `true` by default. `ParseConfig` rejects a `scan_interval_seconds` value of zero or less. Container processes are monitored only when `include_container_processes` is true.

## Code and configuration locations

- `product/process_monitoring/process_monitor.go`
- `product/process_monitoring/config.yaml`
- `product/process_monitoring/process_monitor_test.go`

The default scan interval is 60 seconds.

## Tests and verification

Run `go test ./product/process_monitoring` to verify configuration and process selection behavior.

`TestCheckedInConfigUsesSixtySecondInterval` reads the checked-in `config.yaml` and asserts that the scan interval is 60. `TestParseConfigReadsCheckedInDefault` parses an inline YAML string with a 30-second value. It checks parsing and does not read the checked-in file, so it does not reflect the current default.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet. No measured CPU or performance impact of the 60-second default is established.
