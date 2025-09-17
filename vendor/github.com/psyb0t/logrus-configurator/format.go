package logrusconfigurator

import (
	"fmt"
	"path"
	"runtime"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type format string

const (
	formatJSON format = "json"
	formatText format = "text"
)

func getLogrusFormat(format format) (logrus.Formatter, error) { //nolint:ireturn
	callerPrettyfier := func(f *runtime.Frame) (string, string) {
		filename := path.Base(f.File)

		return fmt.Sprintf("%s()", f.Function),
			fmt.Sprintf("%s:%d", filename, f.Line)
	}

	switch format {
	case formatJSON:
		return &logrus.JSONFormatter{
			CallerPrettyfier: callerPrettyfier,
		}, nil
	case formatText:
		return &logrus.TextFormatter{
			CallerPrettyfier: callerPrettyfier,
		}, nil
	default:
		return nil, errors.Wrap(errInvalidLogFormat, string(format))
	}
}

func setFormat(fmt format) error {
	logrusFormatter, err := getLogrusFormat(fmt)
	if err != nil {
		return err
	}

	logrus.SetFormatter(logrusFormatter)

	return nil
}
