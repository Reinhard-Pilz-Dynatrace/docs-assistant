# Log collection Developer Guide

## Behavior and ownership

The log-collection behavior is owned by the `product/log_collection` package.

`ShouldCollect` returns false when collection is disabled or the file is larger than `max_file_size_mb`. Rotated files are collected only when `include_rotated_files` is true. `ParseConfig` rejects a `max_file_size_mb` value of zero or less.

## Code and configuration locations

- `product/log_collection/log_collector.go`
- `product/log_collection/config.yaml`
- `product/log_collection/log_collector_test.go`

The checked-in defaults are `enabled: true`, `max_file_size_mb: 100`, and `include_rotated_files: false`.

## Tests and verification

Run `go test ./product/log_collection` to verify configuration and file selection behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
