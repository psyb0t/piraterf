package gonfiguration

import "errors"

var (
	ErrInvalidEnvVar        = errors.New("dafuq? we've got a shitty environment variable")
	ErrTargetNotPointer     = errors.New("yo, the destination ain't a pointer")
	ErrDestinationNotStruct = errors.New("what the hell? expected a struct, but this ain't one")
	ErrUnsupportedFieldType = errors.New("wtf.. Unsupported field type")
)
