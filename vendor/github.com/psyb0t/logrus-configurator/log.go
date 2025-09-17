package logrusconfigurator

import (
	"io"

	"github.com/pkg/errors"
	"github.com/psyb0t/gonfiguration"
	"github.com/sirupsen/logrus"
)

const (
	configKeyLogLevel  = "LOG_LEVEL"
	configKeyLogFormat = "LOG_FORMAT"
	configKeyLogCaller = "LOG_CALLER"
)

const (
	defaultReportCaller = false
	defaultLevel        = levelInfo
	defaultFormat       = formatText
)

type config struct {
	Level        level  `env:"LOG_LEVEL"`
	Format       format `env:"LOG_FORMAT"`
	ReportCaller bool   `env:"LOG_CALLER"`
}

func (c config) log() {
	logrus.Debugf(
		"logrus-configurator: level: %s, format: %s, reportCaller: %t",
		c.Level,
		c.Format,
		c.ReportCaller,
	)
}

//nolint:gochecknoinits
func init() {
	if err := configure(); err != nil {
		logrus.Panic(err)
	}
}

func configure() error {
	setDefaults()

	c := config{}
	if err := gonfiguration.Parse(&c); err != nil {
		return errors.Wrap(err, "failed to parse log config")
	}

	if err := setLevel(c.Level); err != nil {
		return errors.Wrap(err, "failed to set log level")
	}

	logrus.SetOutput(io.Discard)
	logrus.SetReportCaller(c.ReportCaller)

	if err := setFormat(c.Format); err != nil {
		return errors.Wrap(err, "failed to set log format")
	}

	clearLoggerHooks(logrus.StandardLogger())
	addLoggerDefaultHooks(logrus.StandardLogger())

	c.log()

	return nil
}

func setDefaults() {
	gonfiguration.SetDefaults(map[string]any{
		configKeyLogLevel:  defaultLevel,
		configKeyLogFormat: defaultFormat,
		configKeyLogCaller: defaultReportCaller,
	})
}
