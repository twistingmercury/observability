package observability_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

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
