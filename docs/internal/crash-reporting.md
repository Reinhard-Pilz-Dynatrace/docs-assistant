# Crash reporting Developer Guide

## Behavior and ownership

The crash reporting behavior is owned by the `product/crash_reporting` package.

`ShouldUploadDump` returns false when reporting is disabled or the dump is larger than `max_dump_size_mb`. `ShouldIncludeEnvironment` requires both `enabled` and `include_environment`. `ParseConfig` rejects a `max_dump_size_mb` value outside 1 to 1024.

## Code and configuration locations

- `product/crash_reporting/crash_reporter.go`
- `product/crash_reporting/config.yaml`
- `product/crash_reporting/crash_reporter_test.go`

The checked-in defaults are `enabled: true`, `max_dump_size_mb: 64`, and `include_environment: false`.

## Tests and verification

Run `go test ./product/crash_reporting` to verify configuration and behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
