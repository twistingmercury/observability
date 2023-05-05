package observability_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability"
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

func TestNewGrpcConnection(t *testing.T) {
	setupTestSvr()
	defer svr.Stop()

	conn, err := observability.NewGrpcConnection("bufnet", insecure.NewCredentials(), false)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}
