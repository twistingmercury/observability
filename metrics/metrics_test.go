package metrics_test

import (
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/metrics"
	"github.com/twistingmercury/observability/testTools"
	"testing"
)

func TestMetrics(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logger.Initialize(logBuf, logrus.DebugLevel, &logrus.TextFormatter{})

	t.Run("NewUpDownCounter_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = metrics.NewUpDownCounter("my-counter", "does stuff")
		})
	})
	t.Run("NewCounter_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = metrics.NewCounter("my-counter", "does stuff")
		})
	})
	t.Run("NewHistogram_panics", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = metrics.NewHistogram("my-counter", "does stuff")
		})
	})

	ctx := context.Background()
	conn, err := testTools.DialContext(ctx)
	assert.NoError(t, err)

	shutdown, err := metrics.Initialize("unit.test", conn)
	assert.NoError(t, err)

	defer func() {
		testTools.Reset(ctx)
		_ = shutdown(ctx)
		conn.Close()
	}()

	upDownCounter, err := metrics.NewUpDownCounter("test_up_down_counter", "test up down counter")
	assert.NoError(t, err)
	assert.NotNil(t, upDownCounter)

	counter, err := metrics.NewCounter("test_up_down_counter", "test up down counter")
	assert.NoError(t, err)
	assert.NotNil(t, counter)

	histogram, err := metrics.NewHistogram("test_histogram", "test histogram")
	assert.NoError(t, err)
	assert.NotNil(t, histogram)

	_ = shutdown(ctx)
}

func TestErrors(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logger.Initialize(logBuf, logrus.DebugLevel, &logrus.JSONFormatter{})

	_, err := metrics.Initialize("unit.test", nil)
	assert.Error(t, err)

	ctx := context.Background()
	conn, err := testTools.DialContext(ctx)
	assert.NoError(t, err)
	_, err = metrics.Initialize("", conn)
}
