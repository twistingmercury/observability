package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/twistingmercury/observability/tracer"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware() gin.HandlerFunc {
	if !tracer.IsInitialized() {
		logrus.Fatal("tracer.Initialize() must be invoked before using the tracing middleware")
	}
	return func(ctx *gin.Context) {
		rCtx, span := tracer.New(ctx.Request.Context(), "inbound-request", trace.SpanKindServer)

		ctx.Request = ctx.Request.Clone(rCtx)

		ctx.Next()

		if ctx.Writer.Status() >= 500 {
			tracer.EndError(span, ctx.Err())
		} else {
			tracer.EndOK(span)
		}
	}
}
