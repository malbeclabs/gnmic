
The `gets` configuration enables periodic gNMI [Get RPC](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#33-retrieving-snapshots-of-state-information) requests to be executed at configurable intervals. This is useful for polling data that doesn't support streaming subscriptions or when point-in-time snapshots are preferred over continuous streams.

Unlike subscriptions which maintain persistent gRPC streams, Gets are stateless requests that retrieve data snapshots at regular intervals. The responses are processed through the same output pipeline as subscriptions, making it easy to send Get data to file, Kafka, or any other configured output.

## Defining Gets

### File-based Gets config

To define periodic Get requests, create a `gets` section in the configuration file:

```yaml
gets:
  # a configurable get name
  get-name:
    # list of strings, list of paths to retrieve
    paths: []
    # string, path to be set as the Get Request Prefix
    prefix:
    # string, value to set as the GetRequest Prefix Target
    target:
    # boolean, if true, the GetRequest Prefix Target will be set to
    # the configured target name under section `targets`.
    # does not apply if the previous field `target` is set.
    set-target: # true | false
    # string, case insensitive, one of ALL, CONFIG, STATE, OPERATIONAL
    # specifies what type of data should be returned
    type: ALL
    # string, case insensitive, defines the gNMI encoding to be used
    encoding: JSON
    # list of strings, schema definition modules
    models: []
    # duration, Golang duration format, e.g: 30s, 1m, 5m.
    # specifies how often to execute the Get request
    interval: 30s
    # list of strings, the list of outputs to send data to.
    # If blank, defaults to all outputs
    outputs:
      - output1
      - output2
    # list of strings, the list of event processors to apply to the data
    event-processors:
      - processor1
    # uint32, depth value as per: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-depth.md
    depth: 0
    # list of strings, specific target names this get applies to.
    # If blank, applies to all configured targets
    targets:
      - target1
      - target2
```

### Get config to gNMI GetRequest

Each Get configuration (under `gets:`) results in periodic [`GetRequest`](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#331-the-getrequest-message) messages being sent to the target(s) at the configured interval.

Each path results in a separate path entry in the GetRequest. The response [`GetResponse`](https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#332-the-getresponse-message) is then converted and sent to the configured outputs.

#### Data Types

The `type` field controls what data is returned:

| Type | Description |
|------|-------------|
| `ALL` | Return all data (configuration and state) - default |
| `CONFIG` | Return only configuration data |
| `STATE` | Return only derived state data |
| `OPERATIONAL` | Return only operational state data |

### Examples

#### Basic periodic Get

This example polls interface counters every 30 seconds:

```yaml
gets:
  interface_counters:
    paths:
      - /interfaces/interface/state/counters
    interval: 30s
    type: STATE
    encoding: json
```

#### Get with multiple paths

Poll multiple paths with the same interval:

```yaml
gets:
  system_info:
    paths:
      - /system/state/hostname
      - /system/state/current-datetime
      - /system/memory/state
    interval: 60s
    type: STATE
```

#### Get with prefix

Use a prefix to simplify path definitions:

```yaml
gets:
  port_stats:
    prefix: /interfaces/interface[name=eth0]
    paths:
      - /state/counters/in-octets
      - /state/counters/out-octets
      - /state/oper-status
    interval: 10s
```

#### Configure multiple Gets

```yaml
# part of ~/gnmic.yml config file
gets:
  interface_counters:
    paths:
      - /interfaces/interface/state/counters
    interval: 30s
    type: STATE
    outputs:
      - kafka_telemetry

  system_health:
    paths:
      - /system/memory/state
      - /system/cpus/cpu/state
    interval: 60s
    type: STATE
    outputs:
      - file_out

  config_backup:
    paths:
      - /
    interval: 1h
    type: CONFIG
    outputs:
      - file_config
```

## Binding Gets to Targets

Gets can be associated with specific targets in two ways:

### Method 1: Specify targets in the Get config

Use the `targets` field in the Get configuration to specify which targets this Get applies to:

```yaml
gets:
  interface_counters:
    paths:
      - /interfaces/interface/state/counters
    interval: 30s
    targets:
      - router1.lab.com
      - router2.lab.com
```

### Method 2: Specify Gets in the target config

Associate Gets with targets by listing them in the target's configuration:

```yaml
targets:
  router1.lab.com:
    username: admin
    password: secret
    subscriptions:
      - port_stats
    gets:
      - interface_counters
      - system_health
  router2.lab.com:
    username: gnmi
    password: telemetry
    gets:
      - system_health
```

!!! note
    If a Get configuration has no `targets` specified and no targets reference it via their `gets` field, the Get will be executed against all configured targets.

## Full Configuration Example

The following example shows a complete configuration with Gets, subscriptions, and outputs working together:

```yaml
username: admin
password: admin123
insecure: true

targets:
  router1.lab.com:
    subscriptions:
      - interface_events
    gets:
      - interface_counters
      - system_health
  router2.lab.com:
    subscriptions:
      - interface_events
    gets:
      - system_health

# Streaming subscriptions for real-time events
subscriptions:
  interface_events:
    paths:
      - /interfaces/interface/state/oper-status
    stream-mode: on-change

# Periodic Gets for polling data
gets:
  interface_counters:
    paths:
      - /interfaces/interface/state/counters/in-octets
      - /interfaces/interface/state/counters/out-octets
      - /interfaces/interface/state/counters/in-errors
      - /interfaces/interface/state/counters/out-errors
    interval: 30s
    type: STATE
    outputs:
      - kafka_telemetry

  system_health:
    paths:
      - /system/memory/state
      - /system/cpus/cpu/state/total/instant
    interval: 60s
    type: STATE
    outputs:
      - kafka_telemetry
      - file_metrics

outputs:
  kafka_telemetry:
    type: kafka
    address: kafka.lab.com:9092
    topic: gnmi-telemetry
    format: event

  file_metrics:
    type: file
    filename: /var/log/gnmic/metrics.json
    format: event
```

## Output Format

Get responses are converted to the same format as subscription responses, making them compatible with all outputs. The metadata includes:

- `source`: The target name
- `get-name`: The name of the Get configuration
- `format`: The output format

Example output in event format:

```json
{
  "name": "get-request",
  "timestamp": 1594065873938155916,
  "tags": {
    "source": "router1.lab.com",
    "get-name": "interface_counters",
    "interface_name": "eth0"
  },
  "values": {
    "/interfaces/interface/state/counters/in-octets": 671552,
    "/interfaces/interface/state/counters/out-octets": 370930
  }
}
```

## Comparison: Gets vs Subscriptions

| Feature | Gets | Subscriptions |
|---------|------|---------------|
| Connection | Stateless (new request each interval) | Persistent gRPC stream |
| Data delivery | Polling at intervals | Real-time streaming |
| Use case | Counters, snapshots, config backup | Events, state changes |
| Resource usage | Lower (no persistent connection) | Higher (maintains stream) |
| Data freshness | Up to interval age | Real-time |

## Metrics

The Get poller exposes Prometheus metrics for monitoring:

| Metric | Type | Description |
|--------|------|-------------|
| `gnmic_get_poller_requests_total` | Counter | Total Get requests sent |
| `gnmic_get_poller_request_duration_seconds` | Histogram | Duration of Get requests |
| `gnmic_get_poller_request_errors_total` | Counter | Total Get request errors |
| `gnmic_get_poller_response_notifications_total` | Counter | Total notifications received |

Labels include `target` and `get_name` for filtering.
