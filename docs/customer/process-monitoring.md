# Process monitoring

## Overview

Process monitoring records host processes when the feature is enabled. By default it scans every 60 seconds, which suits hosts where you prefer a less frequent scan cadence.

## Configuration

Configure process monitoring in `product/process_monitoring/config.yaml`.

- `process_monitoring.enabled`: turns process monitoring on or off.
- `process_monitoring.scan_interval_seconds`: the number of seconds between process scans. The value must be greater than zero.
- `process_monitoring.include_container_processes`: set to `true` to include container processes.

## Defaults

Process monitoring is enabled by default. The default scan interval is 60 seconds.

The default scan interval was previously 30 seconds. The change affects the default cadence only and does not change whether process monitoring is enabled. New installations use the 60-second default. To scan more often, set `scan_interval_seconds` to a lower value.

## Limitations

- Container processes are not included by default.
- The longer default interval changes scan cadence only. No measured CPU or performance improvement is claimed.

## Troubleshooting

Troubleshooting guidance is not established yet.
