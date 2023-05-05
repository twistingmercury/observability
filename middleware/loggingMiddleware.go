// Package middleware contains the middleware for the instrumenting incoming HTTP requests.
package middleware

import (
	"fmt"
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
	rCtx, span := tracer.New(c.Request.Context(), "inbound-request")
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

	if cType := c.Request.Header.Get("Content-Type"); len(cType) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.content-Type", Value: cType})
	}

	if cLen := c.Request.Header.Get("Content-Length"); len(cLen) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.content-Length", Value: cLen})
	}

	if aType := c.Request.Header.Get("Accept"); len(aType) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.accept", Value: aType})
	}

	if aEnc := c.Request.Header.Get("Accept-Encoding"); len(aEnc) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.accept-Encoding", Value: aEnc})
	}

	if xRid := c.Request.Header.Get("x-request-id"); len(xRid) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.x-request-id", Value: xRid})
	}

	if rawq := c.Request.URL.RawQuery; len(rawq) > 0 {
		attribs = append(attribs, logger.Attribute{Key: "http.query", Value: rawq})
	}

	if uaMap := parseUserAgent(c.Request.UserAgent()); len(uaMap) > 0 {
		for k, v := range uaMap {
			if len(v) == 0 {
				continue
			}
			attribs = append(attribs, logger.Attribute{Key: k, Value: v})
		}
	}

	logger.InfoWithSpanContext(c.Request.Context(), "inbound-request", attribs...)
	c.Next()
}

func parseUserAgent(rawUserAgent string) (uaMap map[string]string) {
	uaMap = make(map[string]string)
	if len(rawUserAgent) == 0 {
		return //no-op
	}

	ua := useragent.Parse(rawUserAgent)
	uaMap = map[string]string{
		"http.user_agent.browser":    ua.Name,
		"http.user_agent.os":         ua.OS,
		"http.user_agent.os_Version": ua.OSVersion,
		"http.user_agent.URL":        ua.URL,
		"http.user_agent.device":     ua.Device,
	}

	switch {
	case ua.IsChrome():
		uaMap["http.user_agent.browser"] = "chrome"
	case ua.IsSafari():
		uaMap["http.user_agent.browser"] = "safari"
	case ua.IsFirefox():
		uaMap["http.user_agent.browser"] = "firefox"
	case ua.IsOpera():
		uaMap["http.user_agent.browser"] = "opera"
	case ua.IsInternetExplorer():
		uaMap["http.user_agent.browser"] = "internet_explorer"
	case ua.IsEdge():
		uaMap["http.user_agent.browser"] = "edge"
	default:
		uaMap["http.user_agent.browser"] = "unknown"
	}

	uaMap["http.user_agent.browser_version"] = ua.Version

	switch {
	case ua.Mobile:
		uaMap["http.user_agent.type"] = "mobile"
	case ua.Tablet:
		uaMap["http.user_agent.type"] = "tablet"
	case ua.Desktop:
		uaMap["http.user_agent.type"] = "desktop"
	case ua.IsFacebookbot():
		uaMap["http.user_agent.type"] = "facebook_bot"
	case ua.IsTwitterbot():
		uaMap["http.user_agent.type"] = "twitter_bot"
	case ua.IsGooglebot():
		uaMap["http.user_agent.type"] = "google_bot"
	case ua.Bot:
		uaMap["http.user_agent.type"] = "bot"
	default:
		uaMap["http.user_agent.type"] = "unknown"
	}

	return
}
