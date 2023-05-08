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

func TestNewGrpcConnectionWithBlockError(t *testing.T) {
	opts := observability.GrpcConnectionOptions{
		TransportCreds:  insecure.NewCredentials(),
		WaitForConnect:  true,
		WaitTimeSeconds: 1,
		URL:             "localhost:10101",
	}
	_, err := observability.NewGrpcConnection(opts)
	assert.Error(t, err)
}

func TestNewGrpcConnectionNoBlock(t *testing.T) {
	opts := observability.GrpcConnectionOptions{
		TransportCreds:  insecure.NewCredentials(),
		WaitForConnect:  false,
		WaitTimeSeconds: 1,
		URL:             "localhost:10101",
	}
	_, err := observability.NewGrpcConnection(opts)
	assert.NoError(t, err)
}
