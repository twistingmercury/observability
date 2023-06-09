// Package metrics provides a wrapper around the OpenTelemetry metrics API.
package metrics

import (
	"context"
	"errors"
	"fmt"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/observeCfg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
)

var (
	isInitialized bool
	meter         metric.Meter
	reader        sdkMetric.Reader
	exporter      sdkMetric.Exporter
	provider      *sdkMetric.MeterProvider

	namespace string
	attribs   = []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(observeCfg.ServiceName())},
		{Key: "host", Value: attribute.StringValue(observeCfg.HostName())},
		{Key: "container_id", Value: attribute.StringValue(observeCfg.HostName())},
		{Key: "env", Value: attribute.StringValue(observeCfg.Environment())},
		{Key: "service_version", Value: attribute.StringValue(observeCfg.Version())},
		{Key: "build_date", Value: attribute.StringValue(observeCfg.BuildDate())},
		{Key: "commit_hash", Value: attribute.StringValue(observeCfg.CommitHash())}}
)

func reset() {
	isInitialized = false
	_ = exporter.Shutdown(context.Background())
	exporter = nil
	_ = reader.Shutdown(context.Background())
	_ = provider.Shutdown(context.Background())
}

// IsInitialized returns true if the metrics have been successfully initialized.
func IsInitialized() bool {
	return isInitialized
}

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
		isInitialized = false
		return nil, err
	}
	exporter = exp

	res, err := resource.New(ctx, resource.WithAttributes(attribs...))
	if err != nil {
		return nil, err
	}

	reader = sdkMetric.NewPeriodicReader(exporter)
	option := []sdkMetric.Option{
		sdkMetric.WithReader(reader),
		sdkMetric.WithResource(res),
	}

	meterProvider := sdkMetric.NewMeterProvider(option...)
	global.SetMeterProvider(meterProvider)
	meter = global.Meter(
		fmt.Sprintf("%s.%s", namespace, observeCfg.ServiceName()),
		metric.WithInstrumentationVersion(observeCfg.Version()),
		metric.WithInstrumentationAttributes(attribs...),
	)

	logger.Info("metrics initialized")
	isInitialized = true
	return meterProvider.Shutdown, nil
}

// NewUpDownCounter creates a new up/down counter using the given name and description.
func NewUpDownCounter(name, description string) (c metric.Int64UpDownCounter, err error) {
	opt := []metric.Int64UpDownCounterOption{
		metric.WithDescription(description),
		metric.WithUnit("1"),
	}
	fname := fmt.Sprintf("%s.%s.%s", namespace, observeCfg.ServiceName(), name)
	logger.Debug("new up/down counter created", logger.Attribute{Key: "name", Value: fname})
	return meter.Int64UpDownCounter(fname, opt...)
}

// NewCounter creates a new counter using the given name and description.
func NewCounter(name, description string) (c metric.Int64Counter, err error) {
	opt := []metric.Int64CounterOption{
		metric.WithDescription(description),
		metric.WithUnit("1"),
	}
	fname := fmt.Sprintf("%s.%s.%s", namespace, observeCfg.ServiceName(), name)
	logger.Debug("new up/down counter created", logger.Attribute{Key: "name", Value: fname})
	return meter.Int64Counter(fname, opt...)
}

// NewHistogram creates a new histogram using the given name and description.
func NewHistogram(name, description string) (c metric.Float64Histogram, err error) {
	opt := []metric.Float64HistogramOption{
		metric.WithDescription(description),
		metric.WithUnit("1"),
	}

	fname := fmt.Sprintf("%s.%s.%s", namespace, observeCfg.ServiceName(), name)
	logger.Debug("new histogram created", logger.Attribute{Key: "name", Value: fname})
	return meter.Float64Histogram(fname, opt...)
}
