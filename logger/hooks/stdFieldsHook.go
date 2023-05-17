package hooks

import (
	"github.com/sirupsen/logrus"
	"github.com/twistingmercury/observability/observeCfg"
)

const (
	ServiceDataKey     = "service"
	VersionDataKey     = "version"
	CommitHashDataKey  = "commit_hash"
	EnvironmentDataKey = "env"
	BuildDateDataKey   = "build_date"
	HostDataKey        = "host"
)

type stdFieldsHook struct{}

func NewStdFieldsHook() logrus.Hook {
	return &stdFieldsHook{}
}

func (h *stdFieldsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *stdFieldsHook) Fire(entry *logrus.Entry) (err error) {
	entry.Data[ServiceDataKey] = observeCfg.ServiceName()
	entry.Data[VersionDataKey] = observeCfg.Version()
	entry.Data[CommitHashDataKey] = observeCfg.CommitHash()
	entry.Data[EnvironmentDataKey] = observeCfg.Environment()
	entry.Data[BuildDateDataKey] = observeCfg.BuildDate()
	entry.Data[HostDataKey] = observeCfg.HostName()
	return
}
