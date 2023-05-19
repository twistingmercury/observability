// Package logger provides a wrapper around [logrus] to add standard fields to the log entry.
// It will also add the span context to [logrus] entry.
// [logrus]: https://github.com/sirupsen/logrus
package logger

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

// Attribute is a key-value pair that can be added to a logrus message.
type Attribute struct {
	Key   string
	Value interface{}
}

var (
	isInitialized bool
)

// IsInitialized returns true if the logger has been successfully initialized.
func IsInitialized() bool {
	return isInitialized
}

// Initialize sets up the logger with the given log level and formatter.
func Initialize(out io.Writer, level logrus.Level, hooks ...logrus.Hook) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(level)
	logrus.SetOutput(out)
	for _, h := range hooks {
		logrus.AddHook(h)
	}
	isInitialized = true
}

// Debug logs a message at the debug level and any additional fields passed in as attributes.
func Debug(msg string, attribs ...Attribute) {
	logrus.WithFields(
		withFields(attribs...)).
		Debug(msg)
}

// Info logs a message at the info level and any additional fields passed in as attributes.
func Info(msg string, attribs ...Attribute) {
	logrus.WithFields(
		withFields(attribs...)).
		Info(msg)
}

// Warn logs a message at the warn level and any additional fields passed in as attributes.
func Warn(msg string, attribs ...Attribute) {
	logrus.WithFields(
		withFields(attribs...)).
		Warn(msg)
}

// Error logs a message at the error level and any additional fields passed in as attributes.
func Error(err error, msg string, attribs ...Attribute) {
	logrus.WithFields(
		withFields(attribs...)).
		WithError(err).Error(msg)
}

// Fatal logs a message at the fatal level and any additional fields passed in as attributes.
func Fatal(err error, msg string, attribs ...Attribute) {
	logrus.WithFields(
		withFields(attribs...)).
		WithError(err).Fatal(msg)
}

// ==================== logs with context ====================

// DebugWithSpanContext logs a message at the debug level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func DebugWithSpanContext(sCtx context.Context, msg string, attribs ...Attribute) {
	logrus.WithContext(sCtx).
		WithFields(withFields(attribs...)).
		Debug(msg)
}

// InfoWithSpanContext logs a message at the info level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func InfoWithSpanContext(sCtx context.Context, msg string, attribs ...Attribute) {
	logrus.WithContext(sCtx).
		WithFields(withFields(attribs...)).
		Info(msg)
}

// WarnWithSpanContext logs a message at the warn level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func WarnWithSpanContext(sCtx context.Context, msg string, attribs ...Attribute) {
	logrus.WithContext(sCtx).
		WithFields(withFields(attribs...)).
		Warn(msg)
}

// ErrorWithSpanContext logs a message at the error level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func ErrorWithSpanContext(sCtx context.Context, err error, msg string, attribs ...Attribute) {
	logrus.WithContext(sCtx).
		WithFields(withFields(attribs...)).WithError(err).
		Error(msg)
}

// FatalWithSpanContext logs a message at the fatal level with a span context
// and any additional fields passed in as attributes, as well as the trace_id and span_id.
func FatalWithSpanContext(sCtx context.Context, err error, msg string, attribs ...Attribute) {
	logrus.WithContext(sCtx).
		WithFields(withFields(attribs...)).WithError(err).
		Fatal(msg)
}

// withFields adds standard fields to the logs fields.
func withFields(attribs ...Attribute) logrus.Fields {
	newFields := make(logrus.Fields)

	for _, a := range attribs {
		newFields[a.Key] = a.Value
	}

	return newFields
}
