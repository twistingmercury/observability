package hooks

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

const (
	TraceID = "trace_id"
	SpanID  = "span_id"
)

func NewTraceHook() logrus.Hook {
	return &traceHook{}
}

// traceHook is a logrus hook that adds trace information to the log entry.
type traceHook struct{}

func (t *traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t *traceHook) Fire(entry *logrus.Entry) (err error) {
	if entry.Context == nil {
		return
	}

	span := trace.SpanFromContext(entry.Context)
	if !span.IsRecording() {
		return
	}

	entry.Data["dd.trace_id"] = span.SpanContext().TraceID().String()
	entry.Data["dd.span_id"] = span.SpanContext().SpanID().String()
	entry.Data[TraceID] = span.SpanContext().TraceID().String()
	entry.Data[SpanID] = span.SpanContext().SpanID().String()

	return
}
