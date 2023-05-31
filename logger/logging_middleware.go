package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mileusna/useragent"
)

// LoggingMiddleware logs the incoming request and starts the trace.
func LoggingMiddleware() gin.HandlerFunc {
	if !IsInitialized() {
		logrus.Fatal("logger.Initialize() must be invoked before using the logging middleware")
	}
	return func(ctx *gin.Context) {
		attribs := []Attribute{
			{Key: "http.method", Value: ctx.Request.Method},
			{Key: "http.path", Value: ctx.Request.URL.Path},
			{Key: "http.remoteAddr", Value: ctx.Request.RemoteAddr},
		}

		if rawq := ctx.Request.URL.RawQuery; len(rawq) > 0 {
			attribs = append(attribs, Attribute{Key: "http.query", Value: rawq})
		}

		hd := ParseHeaders(ctx.Request.Header)
		attribs = append(attribs, hd...)

		ua := ParseUserAgent(ctx.Request.UserAgent())
		attribs = append(attribs, ua...)

		Info("inbound-request", attribs...)
		ctx.Next()
	}
}

// ParseHeaders parses the headers and returns a map of attributes.
func ParseHeaders(headers map[string][]string) (hdrMap []Attribute) {
	hdrMap = make([]Attribute, 0, len(headers))
	for k, v := range headers {
		hdrMap = append(hdrMap, Attribute{Key: strings.ToLower(k), Value: v})
	}
	return
}

// ParseUserAgent parses the user agent string and returns a map of attributes.
func ParseUserAgent(rawUserAgent string) (uaMap []Attribute) {
	if len(rawUserAgent) == 0 {
		return //no-op
	}

	ua := useragent.Parse(rawUserAgent)
	uaMap = []Attribute{
		{Key: "http.user_agent.os", Value: ua.OS},
		{Key: "http.user_agent.os_Version", Value: ua.OSVersion},
		{Key: "http.user_agent.URL", Value: ua.URL},
		{Key: "http.user_agent.device", Value: ua.Device},
	}

	switch {
	case ua.Mobile:
		uaMap = append(uaMap, Attribute{Key: "http.user_agent.type", Value: "mobile"})
	case ua.Tablet:
		uaMap = append(uaMap, Attribute{Key: "http.user_agent.type", Value: "tablet"})
	case ua.Desktop:
		uaMap = append(uaMap, Attribute{Key: "http.user_agent.type", Value: "desktop"})
	}

	if ua.Mobile || ua.Tablet || ua.Desktop {
		switch {
		case ua.IsChrome():
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "chrome"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsSafari():
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "safari"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsFirefox():
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "firefox"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsOpera():
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "opera"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsInternetExplorer() || strings.Contains(rawUserAgent, "Trident"):
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "internet_explorer"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		case ua.IsEdge():
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "edge"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ua.Version})
		default:
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: ""})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser_version", Value: ""})
		}
	}

	if ua.Bot {
		switch {
		case ua.IsGooglebot():
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.type", Value: "google_bot"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "googlebot"})
		default:
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.type", Value: "bot"})
			uaMap = append(uaMap, Attribute{Key: "http.user_agent.browser", Value: "unknown"})
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
