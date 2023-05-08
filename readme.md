#  OTEL Observability Wrappers

This repository contains a set of wrappers for the OpenTelemetry API that provides a consistent 
interface for instrumenting applications for observability using [OpenTelemetry](https://opentelemetry.io/docs/).

## Installation 

```bash
go get -u github.com/twistingmercury/observability
```

## Agents

At this time the Go Open Telemetry framework has not implemented logs. for more information, 
see [Statuses and Releases](https://opentelemetry.io/docs/instrumentation/go/#status-and-releases). This means that the
only way to get logs from your application is to use another agent
in addition to the OTEL Collector:

* Github: [open-telemetry/opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib))
* DockerHub: [otel/opentelemetry-collector-contrib](https://hub.docker.com/r/otel/opentelemetry-collector-contrib)

The extra agent used in developing this package was [Vector](https://vector.dev/).

Examples configurations used for developing this packate are in the [agent_configs](agent_configs) directory. In there,
you will also find a sample [docker-compose.yml](agent_configs/docker-compose.yaml) file that can be used to run the 
OTEL Collector and Vector agents, along with the service you are developing.

## Usage

To use the wrappers, you will need to initialize each wrapper you intend to use:

```go
    
```