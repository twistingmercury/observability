package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twistingmercury/observability/config"
)

func TestInitialize(t *testing.T) {
	config.Initialize("service", "build", "version", "env", "commit")

	assert.Equal(t, "commit", config.CommitHash())
	assert.Equal(t, "build", config.BuildDate())
	assert.Equal(t, "version", config.Version())
	assert.Equal(t, "service", config.ServiceName())
}

func TestEnvironment(t *testing.T) {
	os.Setenv(config.SvcNameEnvVar, "")
	config.Initialize("service", "build", "version", "env", "commit")

	assert.Equal(t, "env", config.Environment())
}

func TestServiceName(t *testing.T) {
	os.Setenv(config.SvcNameEnvVar, "custom_service")
	config.Initialize("service", "build", "version", "env", "commit")

	assert.Equal(t, "custom_service", config.ServiceName())
}

func TestLogLevel(t *testing.T) {
	os.Setenv(config.LogLevelEnvVar, "info")
	config.Initialize("service", "build", "version", "env", "commit")

	assert.Equal(t, "info", config.LogLevel())
}

func TestTraceEndpoint(t *testing.T) {
	os.Setenv(config.TraceEndpointEnvVar, "http://localhost:1234")
	config.Initialize("service", "build", "version", "env", "commit")

	assert.Equal(t, "http://localhost:1234", config.TraceEndpoint())
}

func TestMetricsEndpoint(t *testing.T) {
	os.Setenv(config.MetricsEndpointEnvVar, "http://localhost:5678")
	config.Initialize("service", "build", "version", "env", "commit")

	assert.Equal(t, "http://localhost:5678", config.MetricsEndpoint())
}

func TestHostName(t *testing.T) {
	config.Initialize("service", "build", "version", "env", "commit")

	hostname, _ := os.Hostname()
	assert.Equal(t, hostname, config.HostName())
}

func TestShowVersion(t *testing.T) {
	config.Initialize("service", "build", "version", "env", "commit")

	assert.False(t, config.ShowVersion())
}

func TestShowHelp(t *testing.T) {
	config.Initialize("service", "build", "version", "env", "commit")

	assert.False(t, config.ShowHelp())
}

func TestShowHelpWithHelpFlag(t *testing.T) {
	os.Args = []string{"cmd", "--help"}
	config.Initialize("service", "build", "version", "env", "commit")
	assert.True(t, config.ShowHelp())
}

func TestShowVersionWithVersionFlag(t *testing.T) {
	os.Args = []string{"cmd", "--version"}
	config.Initialize("service", "build", "version", "env", "commit")
	assert.True(t, config.ShowVersion())
}

func TestLogLevelWithLogLevelFlag(t *testing.T) {
	os.Args = []string{"cmd", "--log-level", "debug"}
	config.Initialize("service", "build", "version", "env", "commit")
	assert.Equal(t, "debug", config.LogLevel())
}

func TestTraceEndpointWithTraceEndpointFlag(t *testing.T) {
	os.Args = []string{"cmd", "--trace-endpoint", "http://localhost:1234"}
	config.Initialize("service", "build", "version", "env", "commit")
	assert.Equal(t, "http://localhost:1234", config.TraceEndpoint())
}

func TestMetricsEndpointWithMetricsEndpointFlag(t *testing.T) {
	os.Args = []string{"cmd", "--metrics-endpoint", "http://localhost:1234"}
	config.Initialize("service", "build", "version", "env", "commit")
	assert.Equal(t, "http://localhost:1234", config.MetricsEndpoint())
}

func TestEnvironmentWithEnvironmentFlag(t *testing.T) {
	os.Args = []string{"cmd", "--environment", "test"}
	config.Initialize("service", "build", "version", "env", "commit")
	assert.Equal(t, "test", config.Environment())
}
