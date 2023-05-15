package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/metrics"
	"go.opentelemetry.io/otel/metric"
	"time"
)

var (
	activeReq metric.Int64UpDownCounter
	totalReq  metric.Int64Counter
	avgReqDur metric.Float64Histogram
)

var metricsReady bool

func MetricsReady() bool {
	return metricsReady
}

// InitializeMetrics initializes the metrics middleware
func InitializeMetrics() error {
	cr, err := metrics.NewUpDownCounter("http.active_requests", "The current number of active requests being served.")
	if err != nil {
		return fmt.Errorf("failed to create active_requests up down counter: %w", err)
	}

	tr, err := metrics.NewCounter("http.total_requests_served", "The total number of requests served.")
	if err != nil {
		return fmt.Errorf("failed to create total_requests_served counter: %w", err)
	}

	ar, err := metrics.NewHistogram("http.request_duration_seconds", "The request duration in seconds.")
	if err != nil {
		return fmt.Errorf("failed to create request_duration_seconds histogram: %w", err)
	}
	activeReq = cr
	totalReq = tr
	avgReqDur = ar
	metricsReady = true
	return nil
}

// MetricsMiddleware records metrics for the request.
func MetricsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !metricsReady {
			logger.Fatal(errors.New("metrics middleware called before metrics initialized"), "metrics middleware called before metrics initialized")
		}

		defer func(s time.Time) {
			activeReq.Add(ctx.Request.Context(), -1)
			avgReqDur.Record(ctx.Request.Context(), float64(time.Since(s).Microseconds()))
		}(time.Now())

		activeReq.Add(ctx.Request.Context(), 1)
		totalReq.Add(ctx.Request.Context(), 1)

		ctx.Next()
	}
}
