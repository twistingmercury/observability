package tracer_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/testTools"
	"github.com/twistingmercury/observability/tracer"
	"testing"
)

func TestTracingMiddleware_Initialize(t *testing.T) {
	tracer.Reset()
	var fatal = false
	logrus.StandardLogger().ExitFunc = func(int) { fatal = true }
	_ = tracer.TracingMiddleware()
	assert.True(t, fatal)
	fatal = false

	conn, err := testTools.DialContext(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	ts, err := tracer.Initialize(conn)
	assert.NoError(t, err)
	assert.NotNil(t, ts)

	tm := tracer.TracingMiddleware()
	assert.NotNil(t, tm)
	assert.False(t, fatal)
}
