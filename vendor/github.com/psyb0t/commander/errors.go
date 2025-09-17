package commander

import "errors"

// Execution errors
var (
	ErrUnexpectedCommand        = errors.New("unexpected command")
	ErrExpectedCommandNotCalled = errors.New("expected command not called")
)

// Process errors
var (
	ErrProcessStartFailed = errors.New("process start failed")
	ErrProcessWaitFailed  = errors.New("process wait failed")
	ErrPipeCreationFailed = errors.New("pipe creation failed")
)

// Command execution errors
var (
	ErrCommandFailed = errors.New("command failed")
)
