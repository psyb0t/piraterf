package gorpitx

import (
	"errors"
)

// Module execution errors.
var (
	ErrUnknownModule = errors.New("unknown module")
	ErrExecuting     = errors.New("RPITX is busy executing another command")
	ErrNotExecuting  = errors.New("RPITX is not executing a command")
)

// Frequency validation errors (still used by utils.go).
var (
	ErrFreqOutOfRange = errors.New("frequency out of RPiTX range")
	ErrFreqPrecision  = errors.New("frequency precision too high")
)

// PI code validation errors (still used by pifmrds.go).
var (
	ErrPIInvalidHex = errors.New("PI code must be valid hex")
)

// PS validation errors (still used by pifmrds.go).
var (
	ErrPSTooLong = errors.New("PS text must be 8 characters or less")
)
