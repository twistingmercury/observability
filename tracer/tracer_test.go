package tracer_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/testTools"
	"github.com/twistingmercury/observability/tracer"
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

	shutdown, err := tracer.Initialize(conn)
	assert.NoError(t, err)
	assert.True(t, tracer.IsInitialized())
	assert.NotNil(t, shutdown)
	defer func() {
		testTools.Reset(ctx)
		_ = shutdown(ctx)
	}()

	attribs := []attribute.KeyValue{
		attribute.String("test", "test"),
		attribute.Int("test_int", 1),
	}

	cCtx, span := tracer.New(ctx, "test_span", trace.SpanKindUnspecified, attribs...)
	defer tracer.EndOK(span)
	assert.NotNil(t, cCtx)
	assert.NotNil(t, span)
	assert.NotEqual(t, testTools.EmptyTraceId(), span.SpanContext().TraceID().String())
	assert.NotEqual(t, testTools.EmptySpanId(), span.SpanContext().SpanID().String())
	defer tracer.EndOK(span)

	dCtx, span := tracer.New(nil, "test_span", trace.SpanKindUnspecified)
	defer tracer.EndOK(span)
	assert.NotNil(t, dCtx)
	assert.NotNil(t, span)
	assert.NotEqual(t, testTools.EmptyTraceId(), span.SpanContext().TraceID().String())
	assert.NotEqual(t, testTools.EmptySpanId(), span.SpanContext().SpanID().String())
	defer tracer.EndError(span, errors.New("test error"))
}
