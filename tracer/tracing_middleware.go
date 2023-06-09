package tracer

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware() gin.HandlerFunc {
	if !IsInitialized() {
		logrus.Fatal("tracer.Initialize() must be invoked before using the tracing middleware")
	}
	return func(ctx *gin.Context) {
		rCtx, span := New(ctx.Request.Context(), "inbound-request", trace.SpanKindServer)

		ctx.Request = ctx.Request.Clone(rCtx)

		ctx.Next()

		if ctx.Writer.Status() >= 500 {
			EndError(span, ctx.Err())
		} else {
			EndOK(span)
		}
	}
}
