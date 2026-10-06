# Trace sampling

## Overview

Trace sampling decides which traces are kept, so that platform engineers can follow requests across services without storing every trace.

## Configuration

Configure trace sampling in `product/trace_sampling/config.yaml`.

- `trace_sampling.enabled`: turns trace sampling on or off.
- `trace_sampling.sample_ratio_percent`: the percentage of traces that are sampled. The value must be between 1 and 100.
- `trace_sampling.propagate_baggage`: when set to `true`, baggage (request-scoped key/value pairs) is propagated between services.

## Defaults

Trace sampling is enabled by default. The default value of `sample_ratio_percent` is 5 percent. Baggage is not propagated by default (`trace_sampling.propagate_baggage: false`).

## Limitations

No platform limitations for trace sampling are established.

## Troubleshooting

Troubleshooting guidance is not established yet.
