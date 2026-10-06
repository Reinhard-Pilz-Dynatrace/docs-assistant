# Trace sampling Developer Guide

## Behavior and ownership

The trace sampling behavior is owned by the `product/trace_sampling` package.

`ShouldSample` returns false when sampling is disabled; otherwise it samples every `100 / sample_ratio_percent`-th trace, with the step rounded down. `ShouldPropagateBaggage` requires both `enabled` and `propagate_baggage`. `ParseConfig` rejects a `sample_ratio_percent` value outside 1 to 100.

## Code and configuration locations

- `product/trace_sampling/trace_sampler.go`
- `product/trace_sampling/config.yaml`
- `product/trace_sampling/trace_sampler_test.go`

The checked-in defaults are `enabled: true`, `sample_ratio_percent: 5`, and `propagate_baggage: false`.

## Tests and verification

Run `go test ./product/trace_sampling` to verify configuration and behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
