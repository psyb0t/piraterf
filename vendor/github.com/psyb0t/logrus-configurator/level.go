package logrusconfigurator

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type level string

const (
	levelTrace level = "trace"
	levelDebug level = "debug"
	levelInfo  level = "info"
	levelWarn  level = "warn"
	levelError level = "error"
	levelFatal level = "fatal"
	levelPanic level = "panic"
)

func getLogrusLevel(lvl level) (logrus.Level, error) {
	parsedLevel, err := logrus.ParseLevel(strings.ToLower(string(lvl)))
	if err != nil {
		return 0, errors.Wrap(errInvalidLogLevel, string(lvl))
	}

	return parsedLevel, nil
}

func setLevel(lvl level) error {
	logrusLevel, err := getLogrusLevel(lvl)
	if err != nil {
		return err
	}

	logrus.SetLevel(logrusLevel)

	return nil
}
