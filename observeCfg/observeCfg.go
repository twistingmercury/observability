// Package observeCfg provides a set of functions to retrieve configuration values required by the observability
// packages. It is intended to be used by the main package of a service. Internally, it uses the `github.com/spf13/pflag`
// and `github.com/spf13/viper` packages to for configuration. Because of this, when creating configuration logic for
// a service, it is recommended to use the same packages to avoid conflicts.
package observeCfg

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type EnvironmentType string

const (
	Dev        EnvironmentType = "dev"
	Stage      EnvironmentType = "stage"
	Production EnvironmentType = "prod"
	Test       EnvironmentType = "test"
	local      EnvironmentType = "localhost"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
)

const (
	MetricsEndpointEnvVar = "METRICS_ENDPOINT"
	TraceEndpointEnvVar   = "TRACE_ENDPOINT"
	LogLevelEnvVar        = "LOG_LEVEL"
	EnvironEnvVar         = "ENVIRONMENT"

	environFlag         = "env"
	versionFlag         = "version"
	helpFlag            = "help"
	logLevelFlag        = "log-level"
	traceEndpointFlag   = "trace-endpoint"
	metricsEndpointFlag = "metrics-endpoint"
)

// ==================== flags ====================
var (
	fEnv = pflag.String(environFlag, "", "Set the environment in which the service is running [ localhost | dev | test | stage | prod ]")
	fVer = pflag.Bool(versionFlag, false, "Display current version information for the app")
	help = pflag.Bool(helpFlag, false, "Display help information")
	fLlv = pflag.String(logLevelFlag, "", "Sets the log level [ debug | info | warn | error | fatal ]")
	fTep = pflag.String(traceEndpointFlag, "", "The host and port of the otel collector where traces are to be sent [<server>:<port>]")
	fMep = pflag.String(metricsEndpointFlag, "", "The host and port of the otel collector where metrics are to be sent [<server>:<port>]")
)

var (
	// build info
	commitHash string
	bDate      string
	version    string
	name       string
	hostName   string

	// observability config
	levelStr  string
	logLevel  logrus.Level
	traceEP   string
	metricsEP string
	environ   string

	environs = fmt.Sprintf("%s%s%s%s%s", Dev, Stage, Production, Test, local)
)

// Initialize sets the build information, and invokes `pflag.Parse()`.
func Initialize(svcName, buildDate, ver, commit string) (err error) {
	switch {
	case len(svcName) == 0:
		return errors.New("arg svcName cannot be empty")
	case len(buildDate) == 0:
		return errors.New("arg buildDate cannot be empty")
	case len(ver) == 0:
		return errors.New("arg ver cannot be empty")
	case len(commit) == 0:
		return errors.New("arg commit cannot be empty")
	}

	commitHash = commit
	bDate = buildDate
	version = ver
	name = svcName

	if err = bindFlags(); err != nil {
		return
	}
	viper.AutomaticEnv()
	parseConfig()
	return validateConfig()
}

func bindFlags() (err error) {
	pflag.Parse()
	if err = viper.BindPFlag(EnvironEnvVar, pflag.Lookup(environFlag)); err != nil {
		return
	}
	if err = viper.BindPFlag(LogLevelEnvVar, pflag.Lookup(logLevelFlag)); err != nil {
		return
	}
	if err = viper.BindPFlag(TraceEndpointEnvVar, pflag.Lookup(traceEndpointFlag)); err != nil {
		return
	}
	if err = viper.BindPFlag(MetricsEndpointEnvVar, pflag.Lookup(metricsEndpointFlag)); err != nil {
		return
	}
	return
}

func parseConfig() {
	ShowHelp()
	ShowHelp()

	hn, _ := os.Hostname()
	hostName = hn

	levelStr = viper.GetString(LogLevelEnvVar)
	traceEP = viper.GetString(TraceEndpointEnvVar)
	metricsEP = viper.GetString(MetricsEndpointEnvVar)
	environ = viper.GetString(EnvironEnvVar)

	// cli overrides env vars
	if len(*fLlv) != 0 {
		levelStr = *fLlv
	}
	if len(*fTep) != 0 {
		traceEP = *fTep
	}
	if len(*fMep) != 0 {
		metricsEP = *fMep
	}
	if len(*fEnv) != 0 {
		environ = *fEnv
	}
}

func validateConfig() error {
	if len(environ) == 0 {
		return errors.New("environment is required")
	}
	if len(levelStr) == 0 {
		return errors.New("log level is required")
	}
	if len(traceEP) == 0 {
		return errors.New("trace endpoint is required")
	}
	if len(metricsEP) == 0 {
		return errors.New("metrics endpoint is required")
	}

	if !strings.Contains(environs, environ) {
		return fmt.Errorf("invalid environment: %s; accepted values are `%s`, `%s`, `%s`, `%s`, and  `%s`",
			environ, Dev, Stage, Production, Test, local)
	}

	ll, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return fmt.Errorf("invalid log level: %s; accepted levels are `%s`, `%s`, `%s`, `%s`, and  `%s`",
			levelStr, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel)
	}
	logLevel = ll

	return nil
}

// CommitHash returns the VCS reference of the build. It is set by the build process.
func CommitHash() string {
	return commitHash
}

// BuildDate returns the date of the build. It is set by the build process.
func BuildDate() string {
	return bDate
}

// Version returns the version of the build. It is set by the build process.
func Version() string {
	return version
}

// Environment returns the environment the svcName is running in. It is set by the environment variable `ENV`
// and can be overridden by the `--env` flag.
func Environment() string {
	return environ
}

// ServiceName returns the name of the svcName.
func ServiceName() string {
	return name
}

// LogLevel returns the log level. It is set by the environment variable `LOG_LEVEL` and can be overridden by the
// `--log-level` flag.
func LogLevel() logrus.Level {
	return logLevel
}

// TraceEndpoint returns the OpenTelemetry endpoint for traces to be sent to. It is set by the environment variable
// `TRACE_ENDPOINT` and can be overridden by the `--trace-endpoint` flag.
func TraceEndpoint() string {
	return traceEP
}

// MetricsEndpoint returns the OpenTelemetry endpoint for metrics to be sent to. It is set by the environment variable
// `METRICS_ENDPOINT` and can be overridden by the `--metrics-endpoint` flag.
func MetricsEndpoint() string {
	return metricsEP
}

// HostName returns the hostname of the machine the svcName is running on.
func HostName() string {
	return hostName
}

// ShowHelp displays the help information and exits if the `--help` flag is set.
func ShowHelp() {
	if !*help {
		return
	}
	pflag.Usage()
	os.Exit(0)
}

// ShowVersion displays the version information and exits if the `--version` flag is set.
func ShowVersion() {
	if !*fVer {
		return
	}
	fmt.Printf("Version: %s, Build Date: %s, Build Commit: %s\n", version, bDate, commitHash)
	os.Exit(0)
}
