<img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
<img src="https://img.shields.io/badge/Apache_Kafka-231F20?style=for-the-badge&logo=apache-kafka&logoColor=white" />
<img src="https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white" />

# Synthetic Payment Generator

Synthetic data is artificial data that is generated from original data and a model that is trained to reproduce the characteristics and structure of the original data. This means that synthetic data and original data should deliver very similar results when undergoing the same statistical analysis.

## Overview

Generate payment data and send it to a Kafka topics. `avro` schema is used for serialization, `payment.avsc` schema is defined in `avro` folder.
Go generate is used to generate the `payment` struct from the `avro` schema.

### Schema

```json
{
    "namespace": "confluent.io.examples.serialization.avro",
    "name": "Payment",
    "type": "record",
    "fields": [
        {"name": "id", "type": "string"},
        {"name": "ts", "type": "long"}, 
        {"name": "date_ts", "type": "string"}, 
        {"name": "destination", "type": "string"},
        {"name": "source", "type": "string"},
        {"name": "currency", "type": "string"},
        {"name": "amount", "type": "double"},
        {"name": "status", "type": "string"} 
    ]
}
```

By default, the synthetic producer generates payments and payments status changes following one of the provided workflows.
The number represents the workflow weight, the generator will pick a workflow randomly based on the weight, being 1 the lowest one.

### Workflows

* Initiated, Failed: 1
* Initiated, Rejected: 2
* Initiated, Validated, Failed: 1  
* Initiated, Validated, Rejected: 1  
* Initiated, Validated, Accounted, Failed: 1
* Initiated, Validated, Accounted, Completed: 9
* Initiated, Validated, Accounted, Canceled: 2  
* Initiated, Validated, Accounted, Rejected: 1

Delay between events is simulated time for event processing and updating the payment status.

Delays in milliseconds:

* Initiated: 200
* Rejected: 2000  
* Failed: 3000  
* Validated: 500  
* Accounted: 1000
* Completed: 2000
* Rejected: 1000

For each generated payment, the worker will pick up a workflow and genereate all status update events following the selected workflow, and it will use the Delay between events to simulate latency between status update events. All the workflow items are executed in parallel with the corresponding delay. 

## Kafka Topics

One topic is used for each payment status update, the producer will try to create the topics if they don't exist:

* payment-initiated
* payment-completed
* payment-failed
* payment-canceled
* payment-validated
* payment-accounted
* payment-rejected
  
## Configuration

### Kafka configuration

The following environment variables are used to configure the producer:

* `KAFKA_BOOTSTRAP_SERVERS`: Kafka bootstrap servers.
* `KAFKA_SASL_USERNAME`: Kafka SASL username.
* `KAFKA_SASL_PASSWORD`: Kafka SASL password.

### Schema registry configuration

The following environment variables are used to configure the producer:

* `SCHEMA_REGISTRY_ENDPOINT`: Schema registry endpoint.
* `SCHEMA_REGISTRY_API_KEY`: Schema registry API key.
* `SCHEMA_REGISTRY_API_SECRET`: Schema registry API secret.
  
### Datagen configuration

* `NUM_PAYMENTS`: Number of payments to generate. Default: `100000`
* `NUM_WORKERS`: Number of parallel workers to generate payments status updates. Default: `1000`
* `NUM_SOURCES`: Number of sources to generate payments. Default: `10`. Prefix `bank-` is added to the source name.
* `NUM_DESTINATIONS`: Number of destinations to generate payments. Default: `10`. Prefix `bank-` is added to the destination name.

Delays in milliseconds:

* `DELAY_INITIATED`: Number of milliseconds. Default: `100`
* `DELAY_COMPLETED`: Default: `3000`
* `DELAY_FAILED`: Default: `5000`
* `DELAY_CANCELLED`: Default: `2000`
* `DELAY_REJECTED`: Default: `2000`
* `DELAY_ACCOUNTED`: Default: `1000`
* `DELAY_VALIDATED`: Default: `1000`

### Kafka topics

The generator will try to create the topics on the beggining if they don't exist.

On the other hand, create tbe following topics before running the generator.

```shell
confluent kafka topic create payment-initiated
confluent kafka topic create payment-completed
confluent kafka topic create payment-failed
confluent kafka topic create payment-canceled
confluent kafka topic create payment-validated
confluent kafka topic create payment-accounted
confluent kafka topic create payment-rejected
```

## Run with Docker

* Using environment variables:

```shell
docker run -it --rm \
    --env KAFKA_BOOTSTRAP_SERVERS=<kafka-bootstrap-servers> \
    --env KAFKA_SASL_USERNAME=<kafka-sasl-username> \
    --env KAFKA_SASL_PASSWORD=<kafka-sasl-password> \
    --env SCHEMA_REGISTRY_ENDPOINT=<schema-registry-endpoint> \
    --env SCHEMA_REGISTRY_API_KEY=<schema-registry-api-key> \
    --env SCHEMA_REGISTRY_API_SECRET=<schema-registry-api-secret> \
    --env NUM_PAYMENTS=100000 \
    --env NUM_WORKERS=1000 \ 
    mcolomerc/synth-payment:latest
```

* Using a container environment file:

```shell
docker run --env-file .env mcolomerc/synth-payment:latest
```

or mounting a volume with an `.env` file::

```shell
 docker run -d --name=pgen --mount source=.env,destination=/app/.env,readonly mcolomerc/synth-payment:latest
```

## Run with Kubernetes

* Define secret for Kafka credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kafka-cluster-key
type: kubernetes.io/basic-auth
stringData:
  username: <API_KEY> # required field for kubernetes.io/basic-auth
  password: <API_SECRET> # required field for kubernetes.io/basic-auth
```

* Define secret for Schema Registry credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sr-cluster-key
type: kubernetes.io/basic-auth
stringData:
  username: <API_KEY> # required field for kubernetes.io/basic-auth
  password: <API_SECRET> # required field for kubernetes.io/basic-auth
```

* Run as `KubernetesJob` with environment variables:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: synth-payment
spec:
  restartPolicy: Never
  containers:
    - name: synth-payments
      image: mcolomerc/synth-payment:latest
      env:
        -  name: NUM_PAYMENTS
           value: "100"
        -  name: KAFKA_BOOTSTRAP_SERVER
           value: <BOOTSTRAP_SERVER>:9092
        -  name: KAFKA_SASL_USERNAME
           valueFrom:
              secretKeyRef:
                name: kafka-cluster-key
                key: username 
        -  name: KAFKA_SASL_PASSWORD 
            valueFrom:
              secretKeyRef:
                name: kafka-cluster-key
                key: password
        - name: SCHEMA_REGISTRY_ENDPOINT
          value: <SCHEMA_REGISTRY_URL>
        - name: SCHEMA_REGISTRY_API_KEY
          valueFrom:
            secretKeyRef:
              name: sr-cluster-key
              key: username
        - name: SCHEMA_REGISTRY_API_SECRET
          valueFrom:
            secretKeyRef:
              name: sr-cluster-key
              key: password
```

## Output

Example outputs:

* Workers: 100 

```shell
 Starting producer... [100] payments took 6.345051222s

19:08:09.297 [info] Number of runnable goroutines: 103
19:08:09.298 [info] Alloc = 26 MiB
19:08:09.298 [info]     TotalAlloc = 32 MiB
19:08:09.298 [info]     Sys = 40 MiB
19:08:09.298 [info]     NumGC = 2
```

```shell
 Starting producer... [10000] payments took 3m57.345834924s

19:56:36.785 [info] Number of runnable goroutines: 101
19:56:36.785 [info] Alloc = 45 MiB
19:56:36.785 [info]     TotalAlloc = 348 MiB
19:56:36.785 [info]     Sys = 61 MiB
19:56:36.785 [info]     NumGC = 15
```

* Workers: 1000

```shell
 Starting producer... [10000] payments took 27.483902633s

20:00:50.305 [info] Number of runnable goroutines: 1003
20:00:50.306 [info] Alloc = 36 MiB
20:00:50.306 [info]     TotalAlloc = 259 MiB
20:00:50.306 [info]     Sys = 78 MiB
20:00:50.306 [info]     NumGC = 11
```
