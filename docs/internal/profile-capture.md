# Profile capture Developer Guide

## Behavior and ownership

The profile capture behavior is owned by the `product/profile_capture` package.

`ShouldStopProfile` returns true once a profile has run for `max_profile_seconds` while capture is enabled. `ShouldCaptureNativeFrames` requires both `enabled` and `capture_native_frames`. `ParseConfig` rejects a `max_profile_seconds` value outside 1 to 3600.

## Code and configuration locations

- `product/profile_capture/profile_capturer.go`
- `product/profile_capture/config.yaml`
- `product/profile_capture/profile_capturer_test.go`

The checked-in defaults are `enabled: false`, `max_profile_seconds: 60`, and `capture_native_frames: false`.

## Tests and verification

Run `go test ./product/profile_capture` to verify configuration and behavior. `TestCheckedInConfigDefaults` reads the checked-in `config.yaml` and asserts the defaults.

## Operational notes

Operational diagnostics and platform-specific behavior are not established yet.
