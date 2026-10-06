# Metric export Developer Guide

## Behavior and ownership

The metric export behavior is owned by the `product/metric_export` package.

`ShouldExport` returns true when export is enabled and at least `export_interval_seconds` have passed since the last export. `ShouldIncludeHistograms` requires both `enabled` and `include_histograms`. `ParseConfig` rejects a `export_interval_seconds` value outside 1 to 3600.

## Code and configuration locations

- `product/metric_export/metric_exporter.go`
- `product/metric_export/config.yaml`
- `product/metric_export/metric_exporter_test.go`

The checked-in defaults are `enabled: true`, `export_interval_seconds: 15`, and `include_histograms: true`.

## Tests and verification

Run `go test ./product/metric_export` to verify configuration and behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
