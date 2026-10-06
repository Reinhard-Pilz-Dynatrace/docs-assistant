# Crash reporting

## Overview

Crash reporting uploads a dump when a monitored process crashes, so that platform engineers can investigate the cause.

## Configuration

Configure crash reporting in `product/crash_reporting/config.yaml`.

- `crash_reporting.enabled`: turns crash reporting on or off.
- `crash_reporting.max_dump_size_mb`: the largest crash dump, in megabytes, that is uploaded; larger dumps are skipped. The value must be between 1 and 1024.
- `crash_reporting.include_environment`: when set to `true`, environment variables are attached to the crash report.

## Defaults

Crash reporting is enabled by default. The default value of `max_dump_size_mb` is 64 MB. Environment variables are not attached by default (`crash_reporting.include_environment: false`).

## Limitations

No platform limitations for crash reporting are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
