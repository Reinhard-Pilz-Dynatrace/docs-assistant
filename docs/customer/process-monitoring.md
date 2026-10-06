# Process monitoring

## Overview

Process monitoring records host processes when the feature is enabled. Platform engineers who want fewer scans on busy hosts can use the default 60-second scan interval, or adjust it to suit their environment.

## Configuration

Configure process monitoring in `product/process_monitoring/config.yaml`.

- `process_monitoring.enabled`: turns process monitoring on or off.
- `process_monitoring.scan_interval_seconds`: the number of seconds between process scans. The value must be greater than zero.
- `process_monitoring.include_container_processes`: set to `true` to include container processes.

## Defaults

Process monitoring is enabled by default. The default scan interval is 60 seconds. New installations use this default.

This release changes only the default scan cadence, which was previously 30 seconds. It does not change whether process monitoring is enabled.

## Limitations

Container processes are not included by default.

The longer default interval reduces how often scans run. No measured CPU or performance improvement is claimed.

## Troubleshooting

Troubleshooting guidance is not established yet.
