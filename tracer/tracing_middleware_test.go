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

//func TestTracingMiddleware(t *testing.T) {
//	buf := bytes.NewBuffer([]byte{})
//	logger.Initialize(buf, logrus.DebugLevel)
//	assert.True(t, logger.IsInitialized())
//
//	conn, err := testTools.DialContext(context.Background())
//	assert.NoError(t, err)
//	assert.NotNil(t, conn)
//
//	ts, err := tracer.Initialize(conn)
//	assert.NoError(t, err)
//	assert.NotNil(t, ts)
//
//	tm := tracer.TracingMiddleware()
//	assert.NotNil(t, tm)
//
//	gin.SetMode(gin.TestMode)
//	router := gin.New()
//	router.Use(tracer.TracingMiddleware())
//	router.GET("/test", func(c *gin.Context) {
//		c.String(http.StatusOK, "OK")
//	})
//
//	w := httptest.NewRecorder()
//	req, _ := http.NewRequest("GET", "/test", nil)
//	router.ServeHTTP(w, req)
//
//	var logEntry map[string]interface{}
//	err = json.Unmarshal(buf.Bytes(), &logEntry)
//	assert.NoError(t, err)
//
//}
