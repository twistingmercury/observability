package testTools

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
)

const bufSize = 1024 * 1024

var (
	lis          *bufconn.Listener
	svr          *grpc.Server
	emptyTraceId string
	emptySpanId  string
)

func DialContext(ctx context.Context) (*grpc.ClientConn, error) {
	setupTestSvr()
	return grpc.DialContext(ctx,
		"bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func Reset(ctx context.Context) {
	svr.Stop()
	_ = lis.Close()
}

func EmptyTraceId() string {
	return emptyTraceId
}

func EmptySpanId() string {
	return emptySpanId
}

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
