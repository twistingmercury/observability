package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/metrics"
	"github.com/twistingmercury/observability/tracer"
)

// FullMiddlewareChain returns the full middleware chain:
// Tracing -> Logging -> Metrics
func FullMiddlewareChain() gin.HandlersChain {
	return gin.HandlersChain{
		tracer.TracingMiddleware(),
		logger.LoggingMiddleware(),
		metrics.Middleware(),
	}
}
