# Log collection

## Overview

Log collection gathers log files from the host so that platform engineers can review them alongside other monitoring data.

## Configuration

Configure log collection in `product/log_collection/config.yaml`.

- `log_collection.enabled`: turns log collection on or off.
- `log_collection.max_file_size_mb`: the largest log file, in megabytes, that is collected. Larger files are skipped. The value must be greater than zero.
- `log_collection.include_rotated_files`: when set to `true`, rotated log files are collected as well.

## Defaults

Log collection is enabled by default. The default maximum file size is 100 MB. Rotated files are not collected by default (`log_collection.include_rotated_files: false`).

## Limitations

No platform limitations for log collection are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
