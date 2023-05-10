package observeCfg_test

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/twistingmercury/observability/observeCfg"
)

var (
	svcName         = "unittest"
	buildDate       = "2021-01-01T00:00:00Z"
	version         = "v0.0.0"
	commitHash      = "000"
	traceEndpoint   = "localhost:4317"
	metricsEndpoint = "localhost:4327"
	environment     = "localhost"
	logLevel        = logrus.DebugLevel
	originalArgs    = os.Args
)

func setup() {
	os.Setenv(observeCfg.LogLevelEnvVar, "debug")
	os.Setenv(observeCfg.TraceEndpointEnvVar, traceEndpoint)
	os.Setenv(observeCfg.MetricsEndpointEnvVar, metricsEndpoint)
	os.Setenv(observeCfg.EnvironEnvVar, environment)
}

func tearDown() {
	os.Args = originalArgs
	os.Unsetenv(observeCfg.LogLevelEnvVar)
	os.Unsetenv(observeCfg.TraceEndpointEnvVar)
	os.Unsetenv(observeCfg.MetricsEndpointEnvVar)
	os.Unsetenv(observeCfg.EnvironEnvVar)
	viper.Reset()
}

func TestObserveCfg(t *testing.T) {
	t.Run("0-initialize", func(t *testing.T) {
		setup()
		defer tearDown()

		hostName, _ := os.Hostname()

		observeCfg.Initialize(svcName, buildDate, version, commitHash)
		assert.Equal(t, commitHash, observeCfg.CommitHash())
		assert.Equal(t, buildDate, observeCfg.BuildDate())
		assert.Equal(t, version, observeCfg.Version())
		assert.Equal(t, svcName, observeCfg.ServiceName())
		assert.Equal(t, environment, observeCfg.Environment())
		assert.Equal(t, logLevel, observeCfg.LogLevel())
		assert.Equal(t, traceEndpoint, observeCfg.TraceEndpoint())
		assert.Equal(t, metricsEndpoint, observeCfg.MetricsEndpoint())
		assert.Equal(t, hostName, observeCfg.HostName())
	})
	t.Run("1-invalid_log_level", func(t *testing.T) {
		setup()
		defer tearDown()
		os.Args = []string{"cmd"}
		os.Setenv(observeCfg.LogLevelEnvVar, "invalid")

		assert.Panics(t, func() {
			observeCfg.Initialize(svcName, buildDate, version, commitHash)
		})
	})
	t.Run("2-missing_log_level", func(t *testing.T) {
		setup()
		defer tearDown()
		observeCfg.Initialize(svcName, buildDate, version, commitHash)
		os.Unsetenv(observeCfg.LogLevelEnvVar)
		assert.Panics(t, func() {
			observeCfg.Initialize(svcName, buildDate, version, commitHash)
		})
	})
	t.Run("3-invalid_trace_endpoint", func(t *testing.T) {
		setup()
		defer tearDown()
		observeCfg.Initialize(svcName, buildDate, version, commitHash)
		os.Unsetenv(observeCfg.TraceEndpointEnvVar)
		assert.Panics(t, func() {
			observeCfg.Initialize(svcName, buildDate, version, commitHash)
		})
	})
	t.Run("4-invalid_metrics_endpoint", func(t *testing.T) {
		setup()
		defer tearDown()
		observeCfg.Initialize(svcName, buildDate, version, commitHash)
		os.Unsetenv(observeCfg.MetricsEndpointEnvVar)
		assert.Panics(t, func() {
			observeCfg.Initialize(svcName, buildDate, version, commitHash)
		})
	})
	t.Run("5-invalid_environment", func(t *testing.T) {
		setup()
		defer tearDown()
		observeCfg.Initialize(svcName, buildDate, version, commitHash)
		os.Setenv(observeCfg.EnvironEnvVar, "pord")
		assert.Panics(t, func() {
			observeCfg.Initialize(svcName, buildDate, version, commitHash)
		})
	})
	t.Run("6-missing_environment", func(t *testing.T) {
		setup()
		defer tearDown()
		observeCfg.Initialize(svcName, buildDate, version, commitHash)
		os.Unsetenv(observeCfg.EnvironEnvVar)
		assert.Panics(t, func() {
			observeCfg.Initialize(svcName, buildDate, version, commitHash)
		})
	})
	t.Run("7-cli_override", func(t *testing.T) {
		setup()
		defer tearDown()
		os.Args = []string{"cmd",
			"--log-level", "warn",
			"--trace-endpoint", "localhost:1234",
			"--metrics-endpoint", "localhost:5678",
			"--environment", "stage",
		}
		observeCfg.Initialize(svcName, buildDate, version, commitHash)

		assert.Equal(t, logrus.WarnLevel, observeCfg.LogLevel())
		assert.Equal(t, "localhost:1234", observeCfg.TraceEndpoint())
		assert.Equal(t, "localhost:5678", observeCfg.MetricsEndpoint())
		assert.Equal(t, "stage", observeCfg.Environment())
	})
}
