// Package tracer provides a wrapper around OpenTelemetry to add standard fields to the span.
package tracer

import (
	"context"
	"fmt"
	"github.com/twistingmercury/observability/config"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

var tracer trace.Tracer

// Initialize initializes the OpenTelemetry tracing library.
func Initialize(conn *grpc.ClientConn) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName()),
			semconv.ServiceVersionKey.String(config.Version()),
			attribute.String("service.build_date", config.BuildDate()),
			attribute.String("service.commit", config.CommitHash()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	// set global propagator to trace context (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)
	tracer = tracerProvider.Tracer(config.ServiceName())

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

// Attributes represent additional key-value descriptors that can be bound
// to a metric observer or recorder.
var commonAttrs = []attribute.KeyValue{
	semconv.ServiceNameKey.String(config.ServiceName()),
	semconv.ServiceVersionKey.String(config.Version()),
	attribute.String("service.build_date", config.BuildDate()),
	attribute.String("service.commit", config.CommitHash()),
	attribute.String("service.environment", config.Environment()),
	attribute.String("host", config.HostName()),
	attribute.String("container_id", config.HostName()),
}

// New starts a new span with the given name and returns the context and span.
// If spanCtx is nil, context.Background() is used.
// The arg kind is used to set the span kind. The constant trace.SpanKind is defined here: https://pkg.go.dev/go.opentelemetry.io/otel/trace@v1.15.1#SpanKind
func New(spanCtx context.Context, spanName string, kind trace.SpanKind, attributes ...attribute.KeyValue) (ctx context.Context, span trace.Span) {
	if spanCtx == nil {
		spanCtx = context.Background()
	}

	if len(attributes) > 0 {
		commonAttrs = append(commonAttrs, attributes...)
	}

	ctx, span = tracer.Start(
		spanCtx,
		spanName,
		trace.WithSpanKind(kind),
		trace.WithAttributes(commonAttrs...))

	return
}

// EndOK ends the span with a status of "ok".
func EndOK(span trace.Span) {
	span.SetStatus(otelCodes.Ok, "ok")
	span.End()
}

// EndError ends the span with a status of "error".
func EndError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(otelCodes.Error, "error")
	span.End()
}
