// Package middleware contains the middleware for the instrumenting incoming HTTP requests.
package middleware

import (
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
	"github.com/twistingmercury/observability/logger"
	"github.com/twistingmercury/observability/metrics"
	"github.com/twistingmercury/observability/tracer"
	"go.opentelemetry.io/otel/metric"
)

var (
	activeReq metric.Int64UpDownCounter
	totalReq  metric.Int64Counter
	avgReqDur metric.Float64Histogram
)

// Initialize sets up the instrumentation middleware for the application.
func Initialize() error {
	cr, err := metrics.NewUpDownCounter("http.active_requests", "The current number of active requests being served.")
	if err != nil {
		return fmt.Errorf("failed to create active_requests counter: %w", err)
	}

	tr, err := metrics.NewCounter("http.total_requests_served", "The total number of requests served.")
	if err != nil {
		return fmt.Errorf("failed to create total_requests_served counter: %w", err)
	}

	ar, err := metrics.NewHistogram("http.request_duration_seconds", "The request duration in seconds.")
	activeReq = cr
	totalReq = tr
	avgReqDur = ar

	return nil
}

// LogRequest logs the incoming request and starts the trace.
func LogRequest(c *gin.Context) {
	rCtx, span := tracer.New(c.Request.Context(), "inbound-request", trace.SpanKindServer)
	activeReq.Add(rCtx, 1)
	totalReq.Add(rCtx, 1)
	s := time.Now()

	defer func() {
		tracer.EndOK(span)
		activeReq.Add(rCtx, -1)
		avgReqDur.Record(rCtx, time.Since(s).Seconds())
	}()

	c.Request = c.Request.Clone(rCtx)

	attribs := []logger.Attribute{
		{Key: "http.method", Value: c.Request.Method},
		{Key: "http.path", Value: c.Request.URL.Path},
		{Key: "http.remoteAddr", Value: c.Request.RemoteAddr},
	}

	if rawq := c.Request.URL.RawQuery; len(rawq) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.query", Value: rawq})
	}

	for _, v := range ParseHeaders(c.Request.Header) {
		attribs = append(attribs, v)
	}

	for _, v := range ParseUserAgent(c.Request.UserAgent()) {
		attribs = append(attribs, v)
	}

	logger.InfoWithSpanContext(c.Request.Context(), "inbound-request", attribs...)
	c.Next()
}

// ParseHeaders parses the headers and returns a map of attributes.
func ParseHeaders(headers map[string][]string) (headerMap map[string]logger.Attribute) {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}

	headerMap = make(map[string]logger.Attribute)

	for _, k := range keys {
		lck := strings.ToLower(k)
		la := logger.Attribute{Key: fmt.Sprintf("http.header.%s", lck),
			Value: strings.Join(headers[k], "; ")}
		headerMap[lck] = la
	}
	return
}

// ParseUserAgent parses the user agent string and returns a map of attributes.
func ParseUserAgent(rawUserAgent string) (uaMap []logger.Attribute) {
	if len(rawUserAgent) == 0 {
		return //no-op
	}
	uaMap = make([]logger.Attribute, 0)
	ua := useragent.Parse(rawUserAgent)
	uaMap = []logger.Attribute{
		{Key: "http.user_agent.os", Value: ua.OS},
		{Key: "http.user_agent.os_Version", Value: ua.OSVersion},
		{Key: "http.user_agent.URL", Value: ua.URL},
		{Key: "http.user_agent.device", Value: ua.Device},
	}

	switch {
	case ua.Mobile:
		uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.type", Value: "mobile"})
	case ua.Tablet:
		uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.type", Value: "tablet"})
	case ua.Desktop:
		uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.type", Value: "desktop"})
		//case ua.Bot:
		//	uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.type", Value: "bot"})
	}

	if ua.Mobile || ua.Tablet || ua.Desktop {
		switch {
		case ua.IsChrome():
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "chrome"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsSafari():
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "safari"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsFirefox():
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "firefox"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsOpera():
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "opera"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsInternetExplorer() || strings.Contains(rawUserAgent, "Trident"):
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "internet_explorer"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsEdge():
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "edge"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		default:
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: ""})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser_version", Value: ""})
		}
	}

	if ua.Bot {
		switch {
		case ua.IsGooglebot():
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.type", Value: "google_bot"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "googlebot"})
		default:
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.type", Value: "bot"})
			uaMap = append(uaMap, logger.Attribute{Key: "http.user_agent.browser", Value: "unknown"})
		}
	}

	for k, atr := range uaMap {
		s := fmt.Sprintf("%v", atr.Value)
		if len(s) == 0 {
			uaMap[k].Value = "unknown"
		}
	}

	return
}
