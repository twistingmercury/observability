package middleware_test

import (
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/metrics"
	"github.com/twistingmercury/observability/testTools"
	tracing "github.com/twistingmercury/observability/tracer"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/observability/middleware"
)

func TestInitialize(t *testing.T) {
	logBuf := &bytes.Buffer{}
	logger.Initialize(logBuf, logrus.DebugLevel)

	ctx := context.TODO()
	conn, err := testTools.DialContext(ctx)
	assert.NoError(t, err)

	shutdown, err := metrics.Initialize("unit.test", conn)
	assert.NoError(t, err)
	defer func() {
		testTools.Reset(ctx)
		_ = shutdown(ctx)
	}()

	_, err = metrics.Initialize("unit-tests", conn)
	assert.NoError(t, err, "failed to initialize metrics")
	assert.NoError(t, middleware.InitializeMetrics())
	l := middleware.LoggingMiddleware()

	assert.NotNil(t, l)
	m := middleware.MetricsMiddleware()
	assert.NotNil(t, m)
	assert.True(t, middleware.MetricsReady())
}

type testCase struct {
	rawUserAgent string
	expected     []logger.Attribute
}

var testCases = []testCase{
	{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		[]logger.Attribute{
			{"http.user_agent.browser", "chrome"},
			{"http.user_agent.browser_version", "58.0.3029.110"},
			{"http.user_agent.os", "Windows"},
			{"http.user_agent.os_Version", "10.0"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Macintosh; Intel macOS 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.2 Safari/605.1.15",
		[]logger.Attribute{
			{"http.user_agent.browser", "safari"},
			{"http.user_agent.browser_version", "14.1.2"},
			{"http.user_agent.os", "macOS"},
			{"http.user_agent.os_Version", "10.15.7"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Linux; Android 10; SM-A505F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.105 Mobile Safari/537.36",
		[]logger.Attribute{
			{"http.user_agent.browser", "chrome"},
			{"http.user_agent.browser_version", "89.0.4389.105"},
			{"http.user_agent.os", "Android"},
			{"http.user_agent.os_Version", "10"},
			{"http.user_agent.device", "SM-A505F"},
			{"http.user_agent.type", "mobile"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_4_2 like macOS) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1",
		[]logger.Attribute{
			{"http.user_agent.browser", "safari"},
			{"http.user_agent.browser_version", "14.0.3"},
			{"http.user_agent.os", "iOS"},
			{"http.user_agent.os_Version", "14.4.2"},
			{"http.user_agent.device", "iPhone"},
			{"http.user_agent.type", "mobile"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134",
		[]logger.Attribute{
			{"http.user_agent.browser", "edge"},
			{"http.user_agent.browser_version", "17.17134"},
			{"http.user_agent.os", "Windows"},
			{"http.user_agent.os_Version", "10.0"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		[]logger.Attribute{
			{"http.user_agent.browser", "firefox"},
			{"http.user_agent.browser_version", "89.0"},
			{"http.user_agent.os", "Windows"},
			{"http.user_agent.os_Version", "10.0"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 OPR/52.0.2871.99",
		[]logger.Attribute{
			{"http.user_agent.browser", "opera"},
			{"http.user_agent.browser_version", "52.0.2871.99"},
			{"http.user_agent.os", "Windows"},
			{"http.user_agent.os_Version", "10.0"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Windows NT 10.0; Trident/7.0; AS; rv:11.0) like Gecko",
		[]logger.Attribute{
			{"http.user_agent.browser", "internet_explorer"},
			{"http.user_agent.browser_version", "7.0"},
			{"http.user_agent.os", "Windows"},
			{"http.user_agent.os_Version", "10.0"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) CustomBrowser",
		[]logger.Attribute{
			{"http.user_agent.browser", "unknown"},
			{"http.user_agent.browser_version", "unknown"},
			{"http.user_agent.os", "Windows"},
			{"http.user_agent.os_Version", "10.0"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "desktop"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (iPad; CPU OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
		[]logger.Attribute{
			{"http.user_agent.browser", "safari"},
			{"http.user_agent.browser_version", "15.0"},
			{"http.user_agent.os", "iOS"},
			{"http.user_agent.os_Version", "15.0"},
			{"http.user_agent.device", "iPad"},
			{"http.user_agent.type", "tablet"},
			{"http.user_agent.URL", "unknown"},
		},
	},
	{
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		[]logger.Attribute{
			{"http.user_agent.browser", "googlebot"},
			{"http.user_agent.os", "unknown"},
			{"http.user_agent.os_Version", "unknown"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "google_bot"},
			{"http.user_agent.URL", "http://www.google.com/bot.html"},
		},
	},
	{
		"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		[]logger.Attribute{
			{"http.user_agent.browser", "unknown"},
			{"http.user_agent.os", "unknown"},
			{"http.user_agent.os_Version", "unknown"},
			{"http.user_agent.device", "unknown"},
			{"http.user_agent.type", "bot"},
			{"http.user_agent.URL", "http://www.bing.com/bingbot.htm"},
		},
	},
}

func TestParseUserAgent(t *testing.T) {
	for _, tc := range testCases {
		uaMap := middleware.ParseUserAgent(tc.rawUserAgent)
		assert.ElementsMatch(t, tc.expected, uaMap, "ParseUserAgent should return the expected map")
	}

	uaMap := middleware.ParseUserAgent(``)
	assert.Empty(t, uaMap, "ParseUserAgent should return an empty map")
}

func TestParseHeaders(t *testing.T) {

	testValue := []string{"test0", "test1", "test2"}
	expected := strings.Join(testValue, "; ")
	headers := map[string][]string{
		"User-Agent":      testValue,
		"Content-Type":    testValue,
		"Content-Length":  testValue,
		"X-Request-Id":    testValue,
		"Accept-Encoding": testValue,
		"Accept":          testValue,
	}

	hdrMap := middleware.ParseHeaders(headers)

	for _, la := range hdrMap {
		assert.Equal(t, expected, strings.Join(la.Value.([]string), "; "), "ParseHeaders should return the expected map")
	}
}

func TestLogRequestMiddleware(t *testing.T) {

	ctx := context.Background()
	conn, err := testTools.DialContext(ctx)
	assert.NoError(t, err)

	_, err = metrics.Initialize("unit-tests", conn)
	assert.NoError(t, err, "failed to initialize metrics")

	_, err = tracing.Initialize(conn)

	assert.NoError(t, err, "failed to initialize middleware")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "should return OK status")
	assert.Equal(t, "OK", w.Body.String(), "should return OK body")
}
