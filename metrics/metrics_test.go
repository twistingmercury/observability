package metrics_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

const bufSize = 1024 * 1024

var (
	lis *bufconn.Listener
	svr *grpc.Server
)

func setupTestSvr() {
	lis = bufconn.Listen(bufSize)
	svr = grpc.NewServer()
	go func() {
		if err := svr.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestMetrics(t *testing.T) {
	setupTestSvr()
	defer svr.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	shutdown, err := metrics.Initialize("unit.test", conn)
	assert.NoError(t, err)

	_ = shutdown(context.TODO())

	upDownCounter, err := metrics.NewUpDownCounter("test_up_down_counter", "test up down counter")
	assert.NoError(t, err)
	assert.NotNil(t, upDownCounter)

	counter, err := metrics.NewCounter("test_up_down_counter", "test up down counter")
	assert.NoError(t, err)
	assert.NotNil(t, counter)

	histogram, err := metrics.NewHistogram("test_histogram", "test histogram")
	assert.NoError(t, err)
	assert.NotNil(t, histogram)

	_ = shutdown(ctx)
}

func TestErrors(t *testing.T) {
	_, err := metrics.Initialize("unit.test", nil)
	assert.Error(t, err)

	setupTestSvr()
	defer svr.Stop()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	_, err = metrics.Initialize("", conn)
}
