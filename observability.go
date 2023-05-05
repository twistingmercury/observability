package observability

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

// NewGrpcConnection dials the OpenTelemetry collector endpoint. This uses a simple backoff algorithm to retry
// the connection if it fails. Normally, this should only happen if the collector and agent are started at the same,
// when running locally.
func NewGrpcConnection(otelEP string, creds credentials.TransportCredentials, withBlock bool) (conn *grpc.ClientConn, err error) {
	logrus.Debugf("connecting to observability endpoint `grpc://%s`", otelEP)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if withBlock {
		conn, err = grpc.DialContext(ctx, otelEP,
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
		)
	} else {
		conn, err = grpc.DialContext(ctx, otelEP,
			grpc.WithTransportCredentials(creds),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %w", err)
	}

	logrus.Debugf("connected to observability endpoint `grpc://%s`", otelEP)
	return
}
