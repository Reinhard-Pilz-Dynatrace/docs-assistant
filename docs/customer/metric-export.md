# Metric export

## Overview

Metric export sends collected metrics to the analysis backend at a regular cadence, so that platform engineers can chart service health.

## Configuration

Configure metric export in `product/metric_export/config.yaml`.

- `metric_export.enabled`: turns metric export on or off.
- `metric_export.export_interval_seconds`: the number of seconds between metric exports. The value must be between 1 and 3600.
- `metric_export.include_histograms`: when set to `true`, histogram metrics are exported in addition to counters and gauges.

## Defaults

Metric export is enabled by default. The default value of `export_interval_seconds` is 30 seconds. Histograms are exported by default (`metric_export.include_histograms: true`).

The change to a 30-second default affects the default cadence only. It does not change whether metric export is enabled.

## Limitations

No platform limitations for metric export are established. The 30-second default is a change in cadence only; no measured cost saving is claimed.

## Troubleshooting

Troubleshooting guidance is not established yet.
