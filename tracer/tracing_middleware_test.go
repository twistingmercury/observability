package tracer_test

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/testTools"
	"github.com/twistingmercury/observability/tracer"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracingMiddleware_Initialize(t *testing.T) {
	tracer.Reset()
	var fatal = false
	orgExitFunc := logrus.StandardLogger().ExitFunc
	logrus.StandardLogger().ExitFunc = func(int) { fatal = true }
	_ = tracer.TracingMiddleware()
	assert.True(t, fatal)
	fatal = false
	defer func() {
		logrus.StandardLogger().ExitFunc = orgExitFunc
	}()

	conn, err := testTools.DialContext(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	ts, err := tracer.Initialize(conn)
	assert.NoError(t, err)
	assert.NotNil(t, ts)
	defer tracer.Reset()

	tm := tracer.TracingMiddleware()
	assert.NotNil(t, tm)
	assert.False(t, fatal)

	w := httptest.NewRecorder()
	_, e := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	e.Use(tm)
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
