# Periodic Gets Example

This example demonstrates how to use the `gets` feature to poll gNMI data at regular intervals.

## Overview

The `gets` configuration allows you to:

- Poll counter data at configurable intervals (e.g., every 30 seconds)
- Retrieve configuration snapshots for backup purposes
- Combine polling with streaming subscriptions in the same deployment
- Route different Get requests to different outputs

## Configuration Highlights

### Gets Configuration

```yaml
gets:
  interface_counters:
    paths:
      - /interfaces/interface/state/counters/in-octets
      - /interfaces/interface/state/counters/out-octets
    interval: 30s        # Poll every 30 seconds
    type: STATE          # Only retrieve state data
    outputs:
      - kafka_out        # Send to Kafka
      - file_counters    # Also save to file
```

### Binding Gets to Targets

You can bind gets to specific targets in two ways:

1. **In the target config:**
```yaml
targets:
  router1:
    gets:
      - interface_counters
      - system_health
```

2. **In the get config:**
```yaml
gets:
  config_backup:
    targets:
      - router1
```

## Use Cases

| Get Name | Interval | Purpose |
|----------|----------|---------|
| `interface_counters` | 30s | Poll interface statistics |
| `system_health` | 60s | Monitor CPU and memory |
| `config_backup` | 1h | Backup device configuration |

## Running the Example

```bash
gnmic --config gnmic.yaml subscribe
```

The `gets` will run alongside any configured subscriptions.

## Outputs

- **Kafka**: All telemetry data goes to `gnmi-telemetry` topic
- **File (counters)**: `/var/log/gnmic/counters.json`
- **File (health)**: `/var/log/gnmic/health.json`
- **File (config)**: `/var/log/gnmic/config-backup.json`

## Monitoring

Prometheus metrics are available for monitoring Get operations:

- `gnmic_get_poller_requests_total` - Total requests sent
- `gnmic_get_poller_request_duration_seconds` - Request latency
- `gnmic_get_poller_request_errors_total` - Failed requests
