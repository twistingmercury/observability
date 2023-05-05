package config

import (
	flag "github.com/spf13/pflag"
	"os"
)

type EnvironmentType string

const (
	Dev        EnvironmentType = "development"
	Stage      EnvironmentType = "staging"
	Production EnvironmentType = "production"
	Test       EnvironmentType = "test"
	local      EnvironmentType = "dev-localhost"
)

const (
	MetricsEndpointEnvVar = "METRICS_ENDPOINT"
	TraceEndpointEnvVar   = "TRACE_ENDPOINT"
	LogLevelEnvVar        = "LOG_LEVEL"
	SvcNameEnvVar         = "SVC_NAME"
)

// ==================== flags ====================
var (
	fEnv = flag.String("environment", "", "Set the environment the sn is running in (development, staging, production, test)")
	fVer = flag.Bool("version", false, "Display current version information")
	help = flag.Bool("help", false, "Display help information")
	fLlv = flag.String("log-level", "", "Set the log level (debug, info, warn, error, fatal)")
	fOep = flag.String("trace-endpoint", "", "The OpenTelemetry endpoint for traces to be sent to")
	fMep = flag.String("metrics-endpoint", "", "The OpenTelemetry endpoint for metrics to be sent to")
)

var (
	ch = "n/a"
	bd = "n/a"
	sv = "n/a"
	ev = "n/a"
	sn = "n/a"
)

// Initialize sets the build information.
func Initialize(s, b, v, e, c string) {
	flag.Parse()
	ch = c
	bd = b
	sv = v
	ev = e
	sn = s
}

// ==================== getters ====================

// CommitHash returns the VCS reference of the build.
func CommitHash() string {
	return ch
}

// BuildDate returns the date of the build.
func BuildDate() string {
	return bd
}

// Version returns the sv of the build.
func Version() string {
	return sv
}

// Environment returns the environment the sn is running in.
func Environment() string {
	if len(*fEnv) != 0 && *fEnv != "n/a" {
		ev = *fEnv
	}
	return ev
}

// ServiceName returns the name of the sn.
// This value can be overriden by the SVC_NAME environment variable.
func ServiceName() string {
	if n := os.Getenv(SvcNameEnvVar); len(n) != 0 && n != "n/a" {
		sn = n
	}
	return sn
}

// LogLevel returns the log level.
func LogLevel() string {
	if len(*fLlv) != 0 && *fLlv != "n/a" {
		return *fLlv
	}
	return os.Getenv(LogLevelEnvVar)
}

// TraceEndpoint returns the OpenTelemetry endpoint for traces to be sent to.
func TraceEndpoint() string {
	if len(*fOep) != 0 && *fOep != "n/a" {
		return *fOep
	}
	return os.Getenv(TraceEndpointEnvVar)
}

// MetricsEndpoint returns the OpenTelemetry endpoint for metrics to be sent to.
func MetricsEndpoint() string {
	if len(*fMep) != 0 && *fMep != "n/a" {
		return *fMep
	}
	return os.Getenv(MetricsEndpointEnvVar)
}

// HostName returns the hostname of the machine the sn is running on.
func HostName() string {
	name, _ := os.Hostname()
	return name
}

// ShowVersion returns true if the sv flag was set.
func ShowVersion() bool {
	return *fVer
}

// ShowHelp returns true if the help flag was set.
func ShowHelp() bool {
	return *help
}
