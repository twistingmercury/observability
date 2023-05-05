package metrics

import (
	"context"
	"errors"
	"fmt"
	"github.com/twistingmercury/observability/config"
	"github.com/twistingmercury/observability/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
)

var (
	meter     metric.Meter
	namespace string
	attribs   = []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(config.ServiceName())},
		{Key: "host", Value: attribute.StringValue(config.HostName())},
		{Key: "container_id", Value: attribute.StringValue(config.HostName())},
		{Key: "env", Value: attribute.StringValue(config.Environment())},
		{Key: "service_version", Value: attribute.StringValue(config.Version())},
		{Key: "build_date", Value: attribute.StringValue(config.BuildDate())},
		{Key: "commit_hash", Value: attribute.StringValue(config.CommitHash())}}
)

// Initialize sets up the metrics using the given grpc connection and namespace.
func Initialize(ns string, conn *grpc.ClientConn) (func(context context.Context) error, error) {
	if conn == nil {
		return nil, errors.New("failed to create the metrics exporter: the grpc connection is nil")
	}
	if len(ns) == 0 {
		return nil, errors.New("failed to create the metrics exporter: the namespace is empty")
	}

	namespace = ns
	ctx := context.Background()
	exp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithGRPCConn(conn),
	)

	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(attribs...))
	if err != nil {
		return nil, err
	}

	reader := sdkMetric.NewPeriodicReader(exp)
	option := []sdkMetric.Option{
		sdkMetric.WithReader(reader),
		sdkMetric.WithResource(res),
	}

	meterProvider := sdkMetric.NewMeterProvider(option...)
	global.SetMeterProvider(meterProvider)
	meter = global.Meter(
		fmt.Sprintf("%s.%s", namespace, config.ServiceName()),
		metric.WithInstrumentationVersion(config.Version()),
		metric.WithInstrumentationAttributes(attribs...),
	)

	logger.Info("metrics initialized")
	return meterProvider.Shutdown, nil
}

// NewUpDownCounter creates a new up/down counter using the given name and description.
func NewUpDownCounter(name, description string) (metric.Int64UpDownCounter, error) {
	opt := []metric.Int64UpDownCounterOption{
		metric.WithDescription(description),
		metric.WithUnit("1"),
	}
	fname := fmt.Sprintf("%s.%s.%s", namespace, config.ServiceName(), name)
	logger.Debug("new up/down counter created", logger.Attribute{Key: "name", Value: fname})
	return meter.Int64UpDownCounter(fname, opt...)
}

// NewCounter creates a new counter using the given name and description.
func NewCounter(name, description string) (metric.Int64Counter, error) {
	opt := []metric.Int64CounterOption{
		metric.WithDescription(description),
		metric.WithUnit("1"),
	}
	fname := fmt.Sprintf("%s.%s.%s", namespace, config.ServiceName(), name)
	logger.Debug("new up/down counter created", logger.Attribute{Key: "name", Value: fname})
	return meter.Int64Counter(fname, opt...)
}

// NewHistogram creates a new histogram using the given name and description.
func NewHistogram(name, description string) (metric.Float64Histogram, error) {
	opt := []metric.Float64HistogramOption{
		metric.WithDescription(description),
		metric.WithUnit("1"),
	}

	fname := fmt.Sprintf("%s.%s.%s", namespace, config.ServiceName(), name)
	logger.Debug("new histogram created", logger.Attribute{Key: "name", Value: fname})
	return meter.Float64Histogram(fname, opt...)
}
