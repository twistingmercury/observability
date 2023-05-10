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
package main 

iimport (
    "github.com/gin-contrib/requestid"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/twistingmercury/observability/observeCfg"
    "github.com/twistingmercury/observability/logger"
    "github.com/twistingmercury/observability/metrics"
    "github.com/twistingmercury/observability/tracer"

    "os"
    ...
)

const serviceName = "my-service"

var ( // build info will be set during the build process
    buildDate    = "{not set}"
    buildVersion = "{not set}"
    buildCommit  = "{not set}"
)

func main(){
    observeCfg.Initialize(serviceName, buildDate, buildVersion, buildCommit)
	logger.Initialize(os.Stdout, logrus.DebugLevel, &logrus.JSONFormatter{})

	shutdownTracer, err := startTracing()
	if err != nil {
		log.Panic(err, "failed to start tracing")
	}

	shutdownMetrics, err := startMetrics()
	if err != nil {
		log.Panic(err, "failed to start metrics")
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = shutdownMetrics(ctx)
		_ = shutdownTracer(ctx)
		cancel()
	}()
	
	// do stuff...start your service, etc.
	r := gin.New()
	r.Use(requestid.New(), middleware.LogRequest, gin.Recovery())
	r.GET("/api/v1/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"ready": true})
	})

	if r.Run(":8080");err != nil {
		log.Panic(err, "error encountered in the gin.Engine.run func")
	}
}

// a helper to start tracing to declutter the main func
func startTracing() (func(context.Context) error, error) {
	// Initialize the tracing
	tConn, err := observability.NewGrpcConnection(observability.GrpcConnectionOptions{
		URL:             oConf.TraceEndpoint(),
		TransportCreds:  insecure.NewCredentials(),
		WaitTimeSeconds: 10,
		WaitForConnect:  false,
	})
	if err != nil {
		log.Panic(err, "failed to create grpc connection for tracing")
	}
	return tracer.Initialize(tConn)
}

// a helper to start metrics to declutter the main func
func startMetrics() (func(context context.Context) error, error) {
	// Initialize the metrics
	mConn, err := observability.NewGrpcConnection(observability.GrpcConnectionOptions{
		URL:             oConf.MetricsEndpoint(),
		TransportCreds:  insecure.NewCredentials(),
		WaitTimeSeconds: 10,
		WaitForConnect:  false,
	})
	if err != nil {
		log.Panic(err, "failed to create grpc connection for metrics")
	}
	return metrics.Initialize("commsagent", mConn)
}
```
## Configuration

the package observeCfg provides a set of functions to retrieve configuration values required by the other observability \
packages. It is intended to be used by the main package of a service. Internally, it uses the [github.com/spf13/pflag](https://pkg.go.dev/github.com/dvln/viper)
and [github.com/spf13/viper](https://pkg.go.dev/github.com/spf13/pflag) packages to for configuration. Because of this, when creating configuration logic specific to
a service, it is recommended to use the same packages to avoid conflicts.

observeCfg should be the first call when starting your app since all other calls within the observability module rely on this package for configuration:

```go
func main(){
    observeCfg.Initialize(serviceName, buildDate, buildVersion, buildCommit)
	// subsequent observability initializers ...
}
```

## Logger

The logger is a simple wrapper around [github.com/sirupsen/logrus](https://pkg.go.dev/github.com/sirupsen/logrus). It is meant
to ensure consistency in how logs are generated and formatted. It should be initialized right after the observeCfg:

```go
func main(){
    observeCfg.Initialize(serviceName, buildDate, buildVersion, buildCommit)
	logger.Initialize(os.Stdout, logrus.DebugLevel, &logrus.JSONFormatter{})
	// ...
}
```

Typically you will write to `stdout`, typical of apps that are containerized. However, if not containerizing, an [io.Writer](https://pkg.go.dev/io#Writer), is included in the logger package:
```go
// HttpWriter is an interface that defines an io.Writer that writes to an HTTP endpoint.
type HttpWriter interface {
	io.Writer
	IsReady() bool
}
```

If you need to send logs to an HTTP endpoint, you can use this io.Writer instead:
```go
func main(){

	w := logger.NewHttpWriter("http://logging-endpoint")
	logger.Initialize(w, logrus.DebugLevel, &logrus.JSONFormatter{})
	// ...
}
```