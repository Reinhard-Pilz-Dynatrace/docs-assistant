# Error tracking

## Overview

Error tracking records application errors with their stack traces, so that platform engineers can see which failures occur most often.

## Configuration

Configure error tracking in `product/error_tracking/config.yaml`.

- `error_tracking.enabled`: turns error tracking on or off.
- `error_tracking.max_stack_frames`: the number of stack frames that are kept for each error. The value must be between 1 and 500.
- `error_tracking.group_similar_errors`: when set to `true`, errors with similar stack traces are grouped into one entry.

## Defaults

Error tracking is enabled by default. The default value of `max_stack_frames` is 50 frames. Similar errors are not grouped by default (`error_tracking.group_similar_errors: false`).

## Limitations

No platform limitations for error tracking are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
