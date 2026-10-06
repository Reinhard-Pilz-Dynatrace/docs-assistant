# Process monitoring

## Overview

Process monitoring records host processes when the feature is enabled. The default scan cadence is one scan every 60 seconds. This suits platform engineers who want fewer process scans on busy hosts while still seeing running processes.

## Configuration

Configure process monitoring in `product/process_monitoring/config.yaml`.

- `process_monitoring.enabled`: turns process monitoring on or off.
- `process_monitoring.scan_interval_seconds`: the number of seconds between process scans. The value must be greater than zero.
- `process_monitoring.include_container_processes`: when set to `true`, container processes are included in monitoring.

## Defaults

Process monitoring is enabled by default. The default scan interval is 60 seconds. New installations use this default.

## Limitations

Container processes are not included by default.

The 60-second default changes the scan cadence only. It does not change whether process monitoring is enabled. No CPU or performance improvement is claimed.

## Troubleshooting

Troubleshooting guidance is not established yet.
