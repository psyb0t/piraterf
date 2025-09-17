package logrusconfigurator

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

func getStderrHook(w io.Writer) logrus.Hook { //nolint:ireturn
	if w == nil {
		w = os.Stderr
	}

	return &writer.Hook{
		Writer: w,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	}
}

func getStdoutHook(w io.Writer) logrus.Hook { //nolint:ireturn
	if w == nil {
		w = os.Stdout
	}

	return &writer.Hook{
		Writer: w,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
			logrus.TraceLevel,
		},
	}
}

func clearLoggerHooks(logger *logrus.Logger) {
	logger.Hooks = make(logrus.LevelHooks)
}

func addLoggerDefaultHooks(logger *logrus.Logger) {
	addLoggerHooks(
		logger,
		getStderrHook(nil),
		getStdoutHook(nil),
	)
}

func addLoggerHooks(logger *logrus.Logger, hooks ...logrus.Hook) {
	for _, hook := range hooks {
		addLoggerHook(logger, hook)
	}
}

func addLoggerHook(logger *logrus.Logger, hook logrus.Hook) {
	logger.AddHook(hook)
}

func setLoggerHooks(logger *logrus.Logger, hooks ...logrus.Hook) {
	clearLoggerHooks(logger)
	addLoggerHooks(logger, hooks...)
}

// SetHooks sets custom hooks for the standard logger, replacing any existing hooks
func SetHooks(hooks ...logrus.Hook) {
	setLoggerHooks(logrus.StandardLogger(), hooks...)
}

// AddHook adds a single hook to the standard logger
func AddHook(hook logrus.Hook) {
	addLoggerHook(logrus.StandardLogger(), hook)
}
