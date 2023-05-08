package observability

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

type GrpcConnectionOptions struct {
	TransportCreds  credentials.TransportCredentials
	WaitForConnect  bool
	WaitTimeSeconds time.Duration
	URL             string
}

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

// NewGrpcConnection dials the OpenTelemetry collector endpoint. This uses a simple backoff algorithm to retry
// the connection if it fails. Normally, this should only happen if the collector and agent are started at the same,
// when running locally.
func newGrpcConn(ctx context.Context, otelEP string, creds credentials.TransportCredentials) (conn *grpc.ClientConn, err error) {
	logrus.Debugf("connecting to observability endpoint `grpc://%s`", otelEP)

	conn, err = grpc.DialContext(ctx, otelEP, grpc.WithTransportCredentials(creds))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %w", err)
	}

	logrus.Debugf("connected to observability endpoint `grpc://%s` initiated", otelEP)
	return
}

func newGrpcConnWithBlock(ctx context.Context, otelEP string, creds credentials.TransportCredentials) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.DialContext(ctx, otelEP,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock())

	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %w", err)
	}
	return
}
