package logrusconfigurator

import "errors"

var (
	errInvalidLogLevel  = errors.New("invalid log level")
	errInvalidLogFormat = errors.New("invalid log format")
)
