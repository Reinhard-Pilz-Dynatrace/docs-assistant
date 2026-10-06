# Error tracking Developer Guide

## Behavior and ownership

The error tracking behavior is owned by the `product/error_tracking` package.

`ShouldKeepFrame` returns true for frames below `max_stack_frames` while tracking is enabled. `ShouldGroupSimilar` requires both `enabled` and `group_similar_errors`. `ParseConfig` rejects a `max_stack_frames` value outside 1 to 500.

## Code and configuration locations

- `product/error_tracking/error_tracker.go`
- `product/error_tracking/config.yaml`
- `product/error_tracking/error_tracker_test.go`

The checked-in defaults are `enabled: true`, `max_stack_frames: 50`, and `group_similar_errors: false`.

## Tests and verification

Run `go test ./product/error_tracking` to verify configuration and behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
