# Process monitoring Developer Guide

## Behavior and ownership

The process-monitoring behavior is owned by the `product/process_monitoring` package.

## Code and configuration locations

- `product/process_monitoring/process_monitor.go`
- `product/process_monitoring/config.yaml`

The default scan interval is 30 seconds.

## Tests and verification

Run `go test ./product/process_monitoring` to verify configuration and process selection behavior.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
