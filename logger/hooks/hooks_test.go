package hooks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/logger/hooks"
	"github.com/twistingmercury/observability/observeCfg"
	"github.com/twistingmercury/observability/testTools"
	"github.com/twistingmercury/observability/tracer"
	"go.opentelemetry.io/otel/trace"
	"os"
	"testing"
)

var (
	buf bytes.Buffer
)

func setup(t *testing.T) {
	os.Setenv(observeCfg.LogLevelEnvVar, "debug")
	os.Setenv(observeCfg.TraceEndpointEnvVar, "traceEndpoint")
	os.Setenv(observeCfg.MetricsEndpointEnvVar, "metricsEndpoint")
	os.Setenv(observeCfg.EnvironEnvVar, "localhost")
	err := observeCfg.Initialize("unit-tests", "2023-01-01T00:00:00.000", "0.0.0", "abcd0123")
	assert.NoError(t, err)
}

func tearDown() {
	os.Unsetenv(observeCfg.LogLevelEnvVar)
	os.Unsetenv(observeCfg.TraceEndpointEnvVar)
	os.Unsetenv(observeCfg.MetricsEndpointEnvVar)
	os.Unsetenv(observeCfg.EnvironEnvVar)
	buf.Reset()
}

func TestStdFieldsHook_FireHook(t *testing.T) {
	type test struct {
		level   logrus.Level
		err     error
		logFunc func(string, ...logger.Attribute)
		errFunc func(error, string, ...logger.Attribute)
	}

	var tests = []test{
		{logrus.DebugLevel, nil, logger.Debug, nil},
		{logrus.InfoLevel, nil, logger.Info, nil},
		{logrus.WarnLevel, nil, logger.Warn, nil},
		{logrus.ErrorLevel, errors.New("test error"), nil, logger.Error},
		{logrus.FatalLevel, errors.New("test fatal"), nil, logger.Fatal},
	}

	logrus.StandardLogger().ExitFunc = func(int) {}
	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			setup(t)
			defer tearDown()
			hook := hooks.NewStdFieldsHook()
			assert.NotNil(t, hook)
			logger.Initialize(&buf, test.level, hook)

			if test.err != nil {
				test.errFunc(test.err, "test message")
			} else {
				test.logFunc("test message")
			}

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			assert.NotEmpty(t, logEntry[hooks.ServiceDataKey], "service data should not be empty")
			assert.NotEmpty(t, logEntry[hooks.EnvironmentDataKey], "environment data should not be empty")
			assert.NotEmpty(t, logEntry[hooks.VersionDataKey], "version data should not be empty")
			assert.NotEmpty(t, logEntry[hooks.HostDataKey], "host data should not be empty")
			assert.NotEmpty(t, logEntry[hooks.BuildDateDataKey], "build date data should not be empty")
			assert.NotEmpty(t, logEntry[hooks.CommitHashDataKey], "commit hash data should not be empty")
		})
	}
}

func TestTraceHook_Fire(t *testing.T) {
	type test struct {
		level   logrus.Level
		err     error
		logFunc func(context.Context, string, ...logger.Attribute)
		errFunc func(context.Context, error, string, ...logger.Attribute)
	}

	var tests = []test{
		{logrus.DebugLevel, nil, logger.DebugWithSpanContext, nil},
		{logrus.InfoLevel, nil, logger.InfoWithSpanContext, nil},
		{logrus.WarnLevel, nil, logger.WarnWithSpanContext, nil},
		{logrus.ErrorLevel, errors.New("test error"), nil, logger.ErrorWithSpanContext},
		{logrus.FatalLevel, errors.New("test fatal"), nil, logger.FatalWithSpanContext},
	}

	logrus.StandardLogger().ExitFunc = func(int) {}
	setup(t)
	defer tearDown()

	conn, err := testTools.DialContext(context.TODO())
	assert.NoError(t, err)

	shutdown, err := tracer.Initialize(conn)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	defer func() {
		testTools.Reset(context.TODO())
		_ = shutdown(context.TODO())
	}()

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			setup(t)
			defer tearDown()
			hook := hooks.NewTraceHook()
			assert.NotNil(t, hook)
			logger.Initialize(&buf, test.level, hook)

			ctx, span := tracer.New(context.Background(), "test_span", trace.SpanKindUnspecified)
			defer span.End()

			if test.err != nil {
				test.errFunc(ctx, test.err, "test message")
			} else {
				test.logFunc(ctx, "test message")
			}

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			eTraceId := span.SpanContext().TraceID().String()
			eSpanId := span.SpanContext().SpanID().String()
			aTraceId := logEntry[hooks.TraceID].(string)
			aSpanId := logEntry[hooks.SpanID].(string)

			assert.Equal(t, eTraceId, aTraceId, "trace id should not be empty")
			assert.Equal(t, eSpanId, aSpanId, "span id should not be empty")
		})
	}
}

func TestTraceHook_Fire_Nil_Empty_Context(t *testing.T) {
	type test struct {
		level   logrus.Level
		err     error
		logFunc func(context.Context, string, ...logger.Attribute)
		errFunc func(context.Context, error, string, ...logger.Attribute)
	}

	var tests = []test{
		{logrus.DebugLevel, nil, logger.DebugWithSpanContext, nil},
		{logrus.InfoLevel, nil, logger.InfoWithSpanContext, nil},
		{logrus.WarnLevel, nil, logger.WarnWithSpanContext, nil},
		{logrus.ErrorLevel, errors.New("test error"), nil, logger.ErrorWithSpanContext},
		{logrus.FatalLevel, errors.New("test fatal"), nil, logger.FatalWithSpanContext},
	}
	logrus.StandardLogger().ExitFunc = func(int) {}
	for _, test := range tests {
		t.Run("nil_context_"+test.level.String(), func(t *testing.T) {
			setup(t)
			defer tearDown()
			hook := hooks.NewTraceHook()
			assert.NotNil(t, hook)
			logger.Initialize(&buf, test.level, hook)

			if test.err != nil {
				test.errFunc(nil, test.err, "test message")
			} else {
				test.logFunc(nil, "test message")
			}

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			assert.Nil(t, logEntry[hooks.TraceID], "trace id should be nil")
			assert.Nil(t, logEntry[hooks.SpanID], "span id should be nil")
		})
	}

	for _, test := range tests {
		t.Run("default_context_"+test.level.String(), func(t *testing.T) {
			setup(t)
			defer tearDown()
			hook := hooks.NewTraceHook()
			assert.NotNil(t, hook)
			logger.Initialize(&buf, test.level, hook)

			if test.err != nil {
				test.errFunc(context.Background(), test.err, "test message")
			} else {
				test.logFunc(context.Background(), "test message")
			}

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			assert.Nil(t, logEntry[hooks.TraceID], "trace id should be nil")
			assert.Nil(t, logEntry[hooks.SpanID], "span id should be nil")
		})
	}
}
