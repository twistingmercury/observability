package tracer_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	tracing "github.com/twistingmercury/observability/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const bufSize = 1024 * 1024

var (
	lis          *bufconn.Listener
	svr          *grpc.Server
	emptyTraceId string
	emptySpanId  string
)

func setupTestSvr() {
	lis = bufconn.Listen(bufSize)
	svr = grpc.NewServer()
	go func() {
		if err := svr.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	ts := new(trace.SpanContext)
	emptyTraceId = ts.TraceID().String()
	emptySpanId = ts.SpanID().String()

}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestTracing(t *testing.T) {
	setupTestSvr()
	defer svr.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	shutdown, err := tracing.Initialize(conn)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)
	defer shutdown(ctx)

	attribs := []attribute.KeyValue{
		attribute.String("test", "test"),
		attribute.Int("test_int", 1),
	}

	cCtx, span := tracing.New(ctx, "test_span", attribs...)
	defer tracing.EndOK(span)
	assert.NotNil(t, cCtx)
	assert.NotNil(t, span)
	assert.NotEqual(t, emptyTraceId, span.SpanContext().TraceID().String())
	assert.NotEqual(t, emptySpanId, span.SpanContext().SpanID().String())
	defer tracing.EndOK(span)

	dCtx, span := tracing.New(nil, "test_span")
	defer tracing.EndOK(span)
	assert.NotNil(t, dCtx)
	assert.NotNil(t, span)
	assert.NotEqual(t, emptyTraceId, span.SpanContext().TraceID().String())
	assert.NotEqual(t, emptySpanId, span.SpanContext().SpanID().String())
	defer tracing.EndError(span, errors.New("test error"))
}
