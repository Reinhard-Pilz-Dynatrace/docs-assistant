# Network monitoring

## Overview

Network monitoring samples network connections so that platform engineers can see which services communicate with each other. Once network monitoring is enabled, DNS queries are recorded for sampled connections by default, so you can see which hostnames your services resolve without any extra configuration.

## Configuration

Configure network monitoring in `product/network_monitoring/config.yaml`.

- `network_monitoring.enabled`: turns network monitoring on or off.
- `network_monitoring.sample_rate_percent`: the percentage of connections that are sampled. The value must be between 1 and 100.
- `network_monitoring.capture_dns_queries`: when set to `true`, DNS queries are recorded for sampled connections. Set it to `false` to stop recording DNS queries.

## Defaults

Network monitoring is disabled by default. When enabled, the default sample rate is 10 percent. DNS queries are captured by default (`network_monitoring.capture_dns_queries: true`), but DNS capture only takes effect when network monitoring is enabled.

## Limitations

No platform limitations for network monitoring are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
