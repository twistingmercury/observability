// Package logger provides a wrapper around [logrus] to add standard fields to the log entry.
// It will also add the span context to [logrus] entry.
// [logrus]: https://github.com/sirupsen/logrus
package logger

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/twistingmercury/observability/observeCfg"
	"go.opentelemetry.io/otel/trace"
)

// Attribute is a key-value pair that can be added to a logrus message.
type Attribute struct {
	Key   string
	Value interface{}
}

// Initialize sets up the logger with the given log level and formatter.
func Initialize(out io.Writer, level logrus.Level, formatter logrus.Formatter) {
	logrus.SetFormatter(formatter)
	logrus.SetLevel(level)
	logrus.SetOutput(out)
}

// ==================== loggers ====================

// Debug logs a message at the debug level and any additional fields passed in as attributes.
func Debug(msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(nil, attribs...)).Debug(msg)
}

// Info logs a message at the info level and any additional fields passed in as attributes.
func Info(msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(nil, attribs...)).Info(msg)
}

// Warn logs a message at the warn level and any additional fields passed in as attributes.
func Warn(msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(nil, attribs...)).Warn(msg)
}

// Error logs a message at the error level and any additional fields passed in as attributes.
func Error(err error, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(nil, attribs...)).WithError(err).Error(msg)
}

// Fatal logs a message at the fatal level and any additional fields passed in as attributes.
func Fatal(err error, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(nil, attribs...)).WithError(err).Fatal(msg)
}

// ==================== logs with context ====================

// DebugWithSpanContext logs a message at the debug level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func DebugWithSpanContext(sCtx context.Context, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(sCtx, attribs...)).Debug(msg)
}

// InfoWithSpanContext logs a message at the info level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func InfoWithSpanContext(sCtx context.Context, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(sCtx, attribs...)).Info(msg)
}

// WarnWithSpanContext logs a message at the warn level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func WarnWithSpanContext(sCtx context.Context, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(sCtx, attribs...)).Warn(msg)
}

// ErrorWithSpanContext logs a message at the error level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func ErrorWithSpanContext(sCtx context.Context, err error, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(sCtx, attribs...)).WithError(err).Error(msg)
}

// FatalWithSpanContext logs a message at the fatal level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func FatalWithSpanContext(sCtx context.Context, err error, msg string, attribs ...Attribute) {
	logrus.WithFields(withFields(sCtx, attribs...)).WithError(err).Fatal(msg)
}

// ==================== helpers ====================
var stdFields = map[string]interface{}{
	"service":      observeCfg.ServiceName(),
	"version":      observeCfg.Version(),
	"commit_hash":  observeCfg.CommitHash(),
	"env":          observeCfg.Environment(),
	"build_date":   observeCfg.BuildDate(),
	"host":         observeCfg.HostName(),
	"container_id": observeCfg.HostName(),
}

// withFields adds standard fields to the logs fields.
func withFields(ctx context.Context, attribs ...Attribute) logrus.Fields {
	newFields := make(logrus.Fields)
	for k, v := range stdFields {
		newFields[k] = v
	}

	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			newFields["dd.trace_id"] = span.SpanContext().TraceID().String()
			newFields["dd.span_id"] = span.SpanContext().SpanID().String()
		}
	}

	for _, a := range attribs {
		newFields[a.Key] = a.Value
	}

	return newFields
}
