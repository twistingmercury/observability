package tracer_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/testTools"
	tracing "github.com/twistingmercury/observability/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"testing"
)

func TestTracing(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logger.Initialize(logBuf, logrus.DebugLevel)

	ctx := context.Background()
	conn, err := testTools.DialContext(ctx)
	assert.NoError(t, err)

	shutdown, err := tracing.Initialize(conn)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	defer func() {
		testTools.Reset(ctx)
		_ = shutdown(ctx)
	}()

	attribs := []attribute.KeyValue{
		attribute.String("test", "test"),
		attribute.Int("test_int", 1),
	}

	cCtx, span := tracing.New(ctx, "test_span", trace.SpanKindUnspecified, attribs...)
	defer tracing.EndOK(span)
	assert.NotNil(t, cCtx)
	assert.NotNil(t, span)
	assert.NotEqual(t, testTools.EmptyTraceId(), span.SpanContext().TraceID().String())
	assert.NotEqual(t, testTools.EmptySpanId(), span.SpanContext().SpanID().String())
	defer tracing.EndOK(span)

	dCtx, span := tracing.New(nil, "test_span", trace.SpanKindUnspecified)
	defer tracing.EndOK(span)
	assert.NotNil(t, dCtx)
	assert.NotNil(t, span)
	assert.NotEqual(t, testTools.EmptyTraceId(), span.SpanContext().TraceID().String())
	assert.NotEqual(t, testTools.EmptySpanId(), span.SpanContext().SpanID().String())
	defer tracing.EndError(span, errors.New("test error"))
}
