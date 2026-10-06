# Profile capture

## Overview

Profile capture records where a process spends its time, so that platform engineers can find slow code paths.

## Configuration

Configure profile capture in `product/profile_capture/config.yaml`.

- `profile_capture.enabled`: turns profile capture on or off.
- `profile_capture.max_profile_seconds`: the longest duration, in seconds, of a single profile. The value must be between 1 and 3600.
- `profile_capture.capture_native_frames`: when set to `true`, native (non-managed) stack frames are captured as well.

## Defaults

Profile capture is disabled by default. The default value of `max_profile_seconds` is 60 seconds. Native frames are not captured by default (`profile_capture.capture_native_frames: false`).

## Limitations

No platform limitations for profile capture are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
