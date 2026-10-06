# Process monitoring

## Overview

Process monitoring records processes when the feature is enabled, including processes running inside containers in the default configuration. This helps platform engineers operating containerized workloads see container processes without changing a setting first. The default scan cadence is one scan every 60 seconds.

## Configuration

Configure process monitoring in `product/process_monitoring/config.yaml`.

- `process_monitoring.enabled`: turns process monitoring on or off.
- `process_monitoring.scan_interval_seconds`: the number of seconds between process scans. The value must be greater than zero.
- `process_monitoring.include_container_processes`: when set to `true`, container processes are included in monitoring. Set it to `false` to exclude container processes.

## Defaults

Process monitoring is enabled by default. The default scan interval is 60 seconds. Container processes are included by default (`process_monitoring.include_container_processes: true`).

Container processes are monitored only while `process_monitoring.enabled` is `true`.

## Limitations

No additional platform limitations for container process monitoring are established.

The 60-second default changes the scan cadence only. It does not change whether process monitoring is enabled. No CPU or performance improvement is claimed.

## Troubleshooting

Troubleshooting guidance is not established yet.
