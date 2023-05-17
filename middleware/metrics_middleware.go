package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/twistingmercury/observability/metrics"
	"go.opentelemetry.io/otel/metric"
	"time"
)

var (
	activeReq metric.Int64UpDownCounter
	totalReq  metric.Int64Counter
	avgReqDur metric.Float64Histogram
)

var isInitialized bool

// InitializeMetrics initializes the metrics middleware
func InitializeMetrics() error {
	isInitialized = false
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
	isInitialized = true
	return nil
}

// MetricsMiddleware records metrics for the request.
func MetricsMiddleware() gin.HandlerFunc {
	if !metrics.IsInitialized() {
		logrus.Fatal("metrics.Initialize() must be called before using the metrics middleware")
	}
	if !isInitialized {
		logrus.Fatal("middleware.InitializeMetrics() must be called before before using the metrics middleware")
	}

	return func(ctx *gin.Context) {
		defer func(s time.Time) {
			activeReq.Add(ctx.Request.Context(), -1)
			avgReqDur.Record(ctx.Request.Context(), float64(time.Since(s).Microseconds()))
		}(time.Now())

		activeReq.Add(ctx.Request.Context(), 1)
		totalReq.Add(ctx.Request.Context(), 1)

		ctx.Next()
	}
}
