package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/twistingmercury/observability/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/logger"
)

const name = "logger_tests"

type loggerTest struct {
	level         logrus.Level
	expectedLevel string
	msg           string
	logFunc       func(string, ...logger.Attribute)
	err           error
	errFunc       func(error, string, ...logger.Attribute)
	attribs       []logger.Attribute
}

type logWithSpanContextTest struct {
	level         logrus.Level
	expectedLevel string
	msg           string
	logFunc       func(context.Context, string, ...logger.Attribute)
	err           error
	errFunc       func(context.Context, error, string, ...logger.Attribute)
	attribs       []logger.Attribute
}

var (
	attribs = []logger.Attribute{
		{"key1", "value1"},
		{"key2", "value2"},
	}

	noContextTests = []loggerTest{
		{logrus.DebugLevel, "debug", "debug message", logger.Debug, nil, nil, nil},
		{logrus.InfoLevel, "info", "info message", logger.Info, nil, nil, nil},
		{logrus.WarnLevel, "warning", "warn message", logger.Warn, nil, nil, nil},
		{logrus.ErrorLevel, "error", "error message", nil, errors.New("error"), logger.Error, nil},
		{logrus.FatalLevel, "fatal", "fatal message", nil, errors.New("fatal"), logger.Fatal, nil},
		{logrus.DebugLevel, "debug", "debug message", logger.Debug, nil, nil, attribs},
		{logrus.InfoLevel, "info", "info message", logger.Info, nil, nil, attribs},
		{logrus.WarnLevel, "warning", "warn message", logger.Warn, nil, nil, attribs},
		{logrus.ErrorLevel, "error", "error message", nil, errors.New("error"), logger.Error, attribs},
		{logrus.FatalLevel, "fatal", "fatal message", nil, errors.New("fatal"), logger.Fatal, attribs},
	}

	withContextTests = []logWithSpanContextTest{
		{logrus.DebugLevel, "debug", "debug message with span context", logger.DebugWithSpanContext, nil, nil, nil},
		{logrus.InfoLevel, "info", "info message with span context", logger.InfoWithSpanContext, nil, nil, nil},
		{logrus.WarnLevel, "warning", "warning message with span context", logger.WarnWithSpanContext, nil, nil, nil},
		{logrus.ErrorLevel, "error", "error message with span context", nil, errors.New("test error"), logger.ErrorWithSpanContext, nil},
		{logrus.FatalLevel, "fatal", "fatal message with span context", nil, errors.New("test fatal"), logger.FatalWithSpanContext, nil},
		{logrus.DebugLevel, "debug", "debug message with span context", logger.DebugWithSpanContext, nil, nil, attribs},
		{logrus.InfoLevel, "info", "info message with span context", logger.InfoWithSpanContext, nil, nil, attribs},
		{logrus.WarnLevel, "warning", "warning message with span context", logger.WarnWithSpanContext, nil, nil, attribs},
		{logrus.ErrorLevel, "error", "error message with span context", nil, errors.New("test error"), logger.ErrorWithSpanContext, attribs},
		{logrus.FatalLevel, "fatal", "fatal message with span context", nil, errors.New("test fatal"), logger.FatalWithSpanContext, attribs},
	}
)

func TestNoTracing(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(int) {}
	for _, test := range noContextTests {
		var buf bytes.Buffer
		logger.Initialize(&buf, test.level, &logrus.JSONFormatter{})

		switch test.level {
		case logrus.ErrorLevel, logrus.FatalLevel:
			test.errFunc(test.err, test.msg, test.attribs...)
		default:
			test.logFunc(test.msg, test.attribs...)
		}

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		assert.NoError(t, err)

		assert.Equal(t, test.msg, logEntry["msg"])
		assert.Equal(t, test.expectedLevel, logEntry["level"])

		if len(test.attribs) > 0 {
			for _, attrib := range test.attribs {
				assert.Equal(t, attrib.Value, logEntry[attrib.Key])
			}
		}

		buf.Reset()
	}
}

func TestWithTracing(t *testing.T) {
	logrus.StandardLogger().ExitFunc = func(int) {}
	for _, test := range withContextTests {
		var buf bytes.Buffer
		logger.Initialize(&buf, test.level, &logrus.JSONFormatter{})
		exporter := tracetest.NewInMemoryExporter()
		bsp := sdktrace.NewBatchSpanProcessor(exporter)
		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(bsp),
		)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		otel.SetTracerProvider(tracerProvider)
		tracer := tracerProvider.Tracer(config.ServiceName())

		ctx, span := tracer.Start(context.Background(), test.expectedLevel)

		switch test.level {
		case logrus.ErrorLevel, logrus.FatalLevel:
			test.errFunc(ctx, test.err, test.msg, test.attribs...)
		default:
			test.logFunc(ctx, test.msg, test.attribs...)
		}

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		assert.NoError(t, err)
		assert.NotEmpty(t, logEntry)

		assert.Equal(t, test.msg, logEntry["msg"])
		assert.Equal(t, test.expectedLevel, logEntry["level"])
		assert.Equal(t, span.SpanContext().TraceID().String(), logEntry["dd.trace_id"])
		assert.Equal(t, span.SpanContext().SpanID().String(), logEntry["dd.span_id"])

		if len(test.attribs) > 0 {
			for _, attrib := range test.attribs {
				assert.Equal(t, attrib.Value, logEntry[attrib.Key])
			}
		}

		_ = tracerProvider.Shutdown(ctx)
		span.End()
		buf.Reset()
	}
}

func TestInitialize(t *testing.T) {
	var buf bytes.Buffer
	logger.Initialize(&buf, logrus.DebugLevel, &logrus.JSONFormatter{})

	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
	assert.IsType(t, &logrus.JSONFormatter{}, logrus.StandardLogger().Formatter)
	assert.Equal(t, &buf, logrus.StandardLogger().Out)
}
