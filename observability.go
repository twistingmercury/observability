package observability

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

// GrpcConnectionOptions are the options for connecting to the OpenTelemetry collector via gRPC.
type GrpcConnectionOptions struct {
	TransportCreds  credentials.TransportCredentials
	WaitForConnect  bool
	WaitTimeSeconds time.Duration
	URL             string
}

// NewGrpcConnection dials the OpenTelemetry collector endpoint.
func NewGrpcConnection(opts GrpcConnectionOptions) (conn *grpc.ClientConn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), opts.WaitTimeSeconds*time.Second)
	defer cancel()

	switch opts.WaitForConnect {
	case true:
		return newGrpcConnWithBlock(ctx, opts.URL, opts.TransportCreds)
	default:
		return newGrpcConn(ctx, opts.URL, opts.TransportCreds)
	}
}

// newGrpcConn dials the OpenTelemetry collector endpoint.
func newGrpcConn(ctx context.Context, otelEP string, creds credentials.TransportCredentials) (conn *grpc.ClientConn, err error) {
	logrus.Debugf("connecting to observability endpoint `grpc://%s`", otelEP)

	conn, err = grpc.DialContext(ctx, otelEP, grpc.WithTransportCredentials(creds))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %w", err)
	}

	logrus.Debugf("connected to observability endpoint `grpc://%s` initiated", otelEP)
	return
}

// newGrpcConnWithBlock dials the OpenTelemetry collector endpoint it will not return until a connection is established or the context is canceled.
func newGrpcConnWithBlock(ctx context.Context, otelEP string, creds credentials.TransportCredentials) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.DialContext(ctx, otelEP,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock())

	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %w", err)
	}
	return
}
