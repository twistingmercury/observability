package middleware

import "github.com/gin-gonic/gin"

// FullMiddlewareChain returns the full middleware chain:
// Tracing -> Logging -> Metrics
func FullMiddlewareChain() gin.HandlersChain {
	return gin.HandlersChain{
		TracingMiddleware(),
		LoggingMiddleware(),
		MetricsMiddleware(),
	}
}
