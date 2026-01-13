`gnmic` supports exporting subscription updates to multiple Apache Kafka brokers/clusters simultaneously

### Configuration sample

A Kafka output can be defined using the below format in `gnmic` config file under `outputs` section:

```yaml
outputs:
  output1:
    # required
    type: kafka 
    # kafka client name. 
    # if left empty, this field is populated with the output name used as output ID (output1 in this example).
    # the full name will be '$(name)-kafka-prod'.
    # If the flag --instance-name is not empty, the full name will be '$(instance-name)-$(name)-kafka-prod.
    # note that each kafka worker (producer) will get client name=$name-$index
    name: ""
    # Comma separated brokers addresses
    address: localhost:9092 
    # Kafka topic name
    topic: telemetry 
    # Kafka topic prefix
    # If supplied, overrides the `topic` key and outputs to a separate topic per source
    # named like `$topic_$subscriptionName_$targetName`. If `source` contains a port number separated with a colon,
    # the colon will be replaced with an underscore due to restrictions on the naming of kafka topics.
    # ex: telemetry_bgp_neighbor_state_device1_6030
    topic-prefix: telemetry
    # starts a sync-producer if set to true.
    sync-producer: false
    # required-acks is used in Produce Requests to tell the broker how many replica acknowledgements
    # it must see before responding. One of `no-response`, `wait-for-local`, `wait-for-all`.
    required-acks: wait-for-local
    # Kafka SASL configuration
    sasl:
      # SASL user name
      user:
      # SASL password
      password:
      # SASL mechanism: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, OAUTHBEARER, and AWS_MSK_IAM are supported
      mechanism:
      # token url for OAUTHBEARER SASL mechanism
      token-url:
      # AWS region for AWS_MSK_IAM mechanism (optional if AWS_REGION env var is set)
      aws-region:
    # tls config
    tls:
      # string, path to the CA certificate file,
      # this will be used to verify the clients certificates when `skip-verify` is false
      ca-file:
      # string, client certificate file.
      cert-file:
      # string, client key file.
      key-file:
      # boolean, if true, the client will not verify the server
      # certificate against the available certificate chain.
      skip-verify: false
    # The total number of times to retry sending a message
    max-retry: 2 
    # Kafka connection timeout
    timeout: 5s 
    # Wait time to reestablish the kafka producer connection after a failure
    recovery-wait-time: 10s 
    # Exported msg format, json, protojson, prototext, proto, event
    format: event 
    # boolean, if true the kafka producer will add a key to 
    # the message written to the broker. The key value is ${source}_${subscription-name}.
    # this is useful for Kafka topics with multiple partitions, it allows to keep messages from the same source and subscription in sequence.
    insert-key: false
    # string, one of `overwrite`, `if-not-present`, ``
    # This field allows populating/changing the value of Prefix.Target in the received message.
    # if set to ``, nothing changes 
    # if set to `overwrite`, the target value is overwritten using the template configured under `target-template`
    # if set to `if-not-present`, the target value is populated only if it is empty, still using the `target-template`
    add-target: 
    # string, a GoTemplate that allows for the customization of the target field in Prefix.Target.
    # it applies only if the previous field `add-target` is not empty.
    # if left empty, it defaults to:
    # {{- if index . "subscription-target" -}}
    # {{ index . "subscription-target" }}
    # {{- else -}}
    # {{ index . "source" | host }}
    # {{- end -}}`
    # which will set the target to the value configured under `subscription.$subscription-name.target` if any,
    # otherwise it will set it to the target name stripped of the port number (if present)
    target-template:
    # boolean, valid only if format is `event`.
    # if true, arrays of events are split and marshaled as JSON objects instead of an array of dicts.
    split-events: false
    # string, a GoTemplate that is executed using the received gNMI message as input.
    # the template execution is the last step before the data is written to the file,
    # First the received message is formatted according to the `format` field above, then the `event-processors` are applied if any
    # then finally the msg-template is executed.
    msg-template:
    # boolean, if true the message timestamp is changed to current time
    override-timestamps: false
    # Number of kafka producers to be created 
    num-workers: 1 
    # (bool) enable debug
    debug: false 
    # (int) number of messages to buffer before being picked up by the workers
    buffer-size: 0
    # (string) enables compression of produced message. One of gzip, snappy, zstd, lz4
    compression-codec: gzip
    # (bool) enables the collection and export (via prometheus) of output specific metrics
    enable-metrics: false 
    # list of processors to apply on the message before writing
    event-processors: 
```

Currently all subscriptions updates (all targets and all subscriptions) are published to the defined topic name unless the `topic-prefix` configuration option is set.

### Kafka Security protocol

Kafka clients can operate with 4 [security protocols](https://kafka.apache.org/24/javadoc/org/apache/kafka/common/security/auth/SecurityProtocol.html), 
their configuration is controlled via both `.tls` and `.sasl` fields under the output config.

**Security Protocol**  | **Description**                           | **Configuration**                       |
-----------------------|-------------------------------------------|-----------------------------------------|
`PLAINTEXT`            | Un-authenticated, non-encrypted channel   | `.tls` and `.sasl` are **NOT** present  |
`SASL_PLAINTEXT`       | SASL authenticated, non-encrypted channel | only `.sasl` is present                 |
`SASL_SSL`             | SASL authenticated, SSL channel           | both `.tls` and `.sasl` are present     |
`SSL`                  | SSL channel                               | only `.tls` is present                  |

#### Security Configuration Examples

=== "PLAINTEXT"
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        # other fields
        # no tls and no sasl fields
    ```
=== "SASL_PLAINTEXT"
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        sasl:
          user: admin
          password: secret
        # other fields
        # no tls field
    ```
=== "SASL_SSL"
    Example1: Without server certificate verification
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        sasl:
          user: admin
          password: secret
        tls:
          skip-verify: true
        # other fields
        # ...
    ```
    Example2: With server certificate verification
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        sasl:
          user: admin
          password: secret
        tls:
          ca-file: /path/to/ca-file
        # other fields
        # ...
    ```
    Example3: With client certificates
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        sasl:
          user: admin
          password: secret
        tls:
          cert-file: /path/to/cert-file
          key-file: /path/to/cert-file
        # other fields
        # ...
    ```
    Example4: With both server certificate verification and client certificates
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        sasl:
          user: admin
          password: secret
        tls:
          cert-file: /path/to/cert-file
          key-file: /path/to/cert-file
          ca-file: /path/to/ca-file
        # other fields
        # ...
    ```
=== "SSL"
    Example1: Without server certificate verification
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        tls:
          skip-verify: true
        # other fields
        # no sasl field
    ```
    Example2: With server certificate verification
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        tls:
          ca-file: /path/to/ca-file
        # other fields
        # no sasl field
    ```
    Example3: With client certificates
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        tls:
          cert-file: /path/to/cert-file
          key-file: /path/to/cert-file
        # other fields
        # no sasl field
    ```
    Example4: With both server certificate verification and client certificates
    ```yaml
    outputs:
      output1:
        type: kafka
        topic: my_kafka_topic
        tls:
          cert-file: /path/to/cert-file
          key-file: /path/to/cert-file
          ca-file: /path/to/ca-file
        # other fields
        # no sasl field
    ```

### AWS MSK IAM Authentication

`gnmic` supports IAM-based authentication for AWS MSK (Managed Streaming for Apache Kafka), including MSK Serverless clusters. This authentication method uses the AWS SDK default credential chain, making it ideal for Kubernetes deployments with IAM Roles for Service Accounts (IRSA).

#### Configuration

To enable AWS MSK IAM authentication, set the SASL mechanism to `AWS_MSK_IAM`:

```yaml
outputs:
  msk-output:
    type: kafka
    address: b-1.mycluster.xxx.kafka.us-east-1.amazonaws.com:9098
    topic: telemetry
    sasl:
      mechanism: AWS_MSK_IAM
      aws-region: us-east-1  # Optional: defaults to AWS_REGION env var
```

**Notes:**

- TLS is automatically enabled when using AWS MSK IAM (required by AWS)
- The `aws-region` field is optional if `AWS_REGION` or `AWS_DEFAULT_REGION` environment variables are set
- User and password fields are not required for AWS MSK IAM

#### Credential Sources

AWS MSK IAM authentication uses the AWS SDK default credential chain, which checks the following sources in order:

1. **Environment variables**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
2. **IAM Roles for Service Accounts (IRSA)**: Automatic credential injection in EKS
3. **EC2 Instance Profile**: Automatic credentials from EC2 metadata service
4. **ECS Task Role**: Automatic credentials for ECS tasks
5. **Shared credentials file**: `~/.aws/credentials`

#### Kubernetes with IRSA

For EKS deployments using IAM Roles for Service Accounts:

1. Create an IAM role with the appropriate MSK permissions
2. Associate the role with your Kubernetes service account
3. Deploy gnmic with the annotated service account

Example IAM policy for MSK access:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kafka-cluster:Connect",
        "kafka-cluster:DescribeCluster"
      ],
      "Resource": "arn:aws:kafka:REGION:ACCOUNT:cluster/CLUSTER_NAME/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kafka-cluster:*Topic*",
        "kafka-cluster:WriteData",
        "kafka-cluster:ReadData"
      ],
      "Resource": "arn:aws:kafka:REGION:ACCOUNT:topic/CLUSTER_NAME/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kafka-cluster:AlterGroup",
        "kafka-cluster:DescribeGroup"
      ],
      "Resource": "arn:aws:kafka:REGION:ACCOUNT:group/CLUSTER_NAME/*"
    }
  ]
}
```

Example Kubernetes service account annotation:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gnmic
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/gnmic-msk-role
```

### Kafka Output Metrics

When a Prometheus server is enabled, `gnmic` kafka output exposes 4 prometheus metrics, 3 Counters and 1 Gauge:

* `number_of_kafka_msgs_sent_success_total`: Number of msgs successfully sent by gnmic kafka output. This Counter is labeled with the kafka producerID
* `number_of_written_kafka_bytes_total`: Number of bytes written by gnmic kafka output. This Counter is labeled with the kafka producerID
* `number_of_kafka_msgs_sent_fail_total`: Number of failed msgs sent by gnmic kafka output. This Counter is labeled with the kafka producerID as well as the failure reason
* `msg_send_duration_ns`: gnmic kafka output send duration in nanoseconds. This Gauge is labeled with the kafka producerID
