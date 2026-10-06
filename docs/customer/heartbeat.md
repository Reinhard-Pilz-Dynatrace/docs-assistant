# Heartbeat

## Overview

The heartbeat tells the backend that an agent is alive, so that platform engineers notice silent hosts quickly.

## Configuration

Configure heartbeat in `product/heartbeat/config.yaml`.

- `heartbeat.enabled`: turns heartbeat on or off.
- `heartbeat.heartbeat_interval_seconds`: the number of seconds between heartbeats. The value must be between 5 and 600.
- `heartbeat.include_host_metadata`: when set to `true`, basic host metadata is attached to each heartbeat.

## Defaults

Heartbeat is enabled by default. The default value of `heartbeat_interval_seconds` is 20 seconds. Host metadata is attached by default (`heartbeat.include_host_metadata: true`).

## Limitations

No platform limitations for heartbeat are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
