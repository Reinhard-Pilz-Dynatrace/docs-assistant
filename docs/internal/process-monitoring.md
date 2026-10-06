# Process monitoring Developer Guide

## Behavior and ownership

The process-monitoring behavior is owned by the `product/process_monitoring` package.

The checked-in default `process_monitoring.scan_interval_seconds` is 60 seconds (changed from 30 in PR #4). `enabled` remains `true` by default. `ParseConfig` rejects a `scan_interval_seconds` value of zero or less.

The checked-in default `process_monitoring.include_container_processes` is `true` (changed in PR #3). `ShouldMonitorProcess` returns false when monitoring is disabled; otherwise it returns true for host processes, and for containerized processes only when `include_container_processes` is true.

The `true` default comes from the checked-in `config.yaml`. If the key is absent from YAML passed to `ParseConfig`, `IncludeContainerProcesses` is the Go zero value, `false`.

## Code and configuration locations

- `product/process_monitoring/process_monitor.go`
- `product/process_monitoring/config.yaml`
- `product/process_monitoring/process_monitor_test.go`

The default scan interval is 60 seconds, and the default for `include_container_processes` is `true`.

## Tests and verification

Run `go test ./product/process_monitoring` to verify configuration and process selection behavior.

`TestCheckedInConfigUsesSixtySecondInterval` and `TestCheckedInConfigIncludesContainerProcesses` both use the `readCheckedInConfig` helper, which reads and parses the checked-in `config.yaml`. They assert that the scan interval is 60 and that `include_container_processes` is true.

`TestParseConfigReadsCheckedInDefault` parses an inline YAML string with a 30-second value and no `include_container_processes` key. It checks parsing and that an absent key yields `false`. It does not read the checked-in file, so it does not reflect the current checked-in defaults.

`TestShouldMonitorProcess` covers host processes, container processes excluded when the option is not set, container processes included when configured, and disabled monitoring.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet. No measured CPU or performance impact of the 60-second default or of including container processes is established.
