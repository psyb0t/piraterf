package commander

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

type Commander interface {
	// Run executes a command and waits for completion
	Run(
		ctx context.Context,
		name string,
		args []string,
		opts ...Option,
	) error

	// Output executes a command and returns stdout, stderr, and error
	Output(
		ctx context.Context,
		name string,
		args []string,
		opts ...Option,
	) (stdout []byte, stderr []byte, err error)

	// CombinedOutput executes a command and returns combined stdout+stderr and error
	CombinedOutput(
		ctx context.Context,
		name string,
		args []string,
		opts ...Option,
	) (output []byte, err error)

	// Start creates a command that can be controlled manually
	Start(
		ctx context.Context,
		name string,
		args []string,
		opts ...Option,
	) (Process, error)
}

func New() Commander { //nolint:ireturn
	return &commander{}
}

type commander struct{}

func (c *commander) Run(
	ctx context.Context,
	name string,
	args []string,
	opts ...Option,
) error {
	options := c.buildOptions(opts...)
	exec := c.newExecutionContext(ctx, name, args, options)
	logrus.Debugf("running command: %s %v", name, args)

	return exec.handleExecutionError(exec.cmd.Run())
}

func (c *commander) Output(
	ctx context.Context,
	name string,
	args []string,
	opts ...Option,
) ([]byte, []byte, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	err := c.runWithOutput(
		ctx,
		name,
		args,
		&stdoutBuf,
		&stderrBuf,
		opts...,
	)

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), err
}

func (c *commander) CombinedOutput(
	ctx context.Context,
	name string,
	args []string,
	opts ...Option,
) ([]byte, error) {
	var combinedBuf bytes.Buffer

	err := c.runWithOutput(
		ctx,
		name,
		args,
		&combinedBuf,
		&combinedBuf,
		opts...,
	)

	return combinedBuf.Bytes(), err
}

func (c *commander) runWithOutput(
	ctx context.Context,
	name string,
	args []string,
	stdoutBuf io.Writer,
	stderrBuf io.Writer,
	opts ...Option,
) error {
	options := c.buildOptions(opts...)

	exec := c.newExecutionContext(
		ctx,
		name,
		args,
		options,
	)

	exec.cmd.Stdout = stdoutBuf
	exec.cmd.Stderr = stderrBuf

	logrus.Debugf("running command for output: %s %v", name, args)

	runErr := exec.cmd.Run()

	logrus.Debug("command output captured")

	return exec.handleExecutionError(runErr)
}

//nolint:ireturn // interface return by design
func (c *commander) Start(
	ctx context.Context,
	name string,
	args []string,
	opts ...Option,
) (Process, error) {
	options := c.buildOptions(opts...)

	execCtx := c.newExecutionContext(
		ctx,
		name,
		args,
		options,
	)

	logrus.Debugf("starting command: %s %v", name, args)

	proc := c.newProcess(execCtx.cmd, execCtx)
	if err := proc.Start(); err != nil {
		return nil, err
	}

	return proc, nil
}

func (c *commander) newProcess(cmd *exec.Cmd, execCtx *executionContext) *process {
	return &process{
		cmd:            cmd,
		execCtx:        execCtx,
		internalStdout: make(chan string),
		internalStderr: make(chan string),
		doneCh:         make(chan struct{}),
		waitCh:         make(chan struct{}),
	}
}

func (c *commander) buildOptions(opts ...Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}

func (c *commander) createCmd(
	ctx context.Context,
	name string,
	args []string,
	opts *Options,
) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	if opts != nil {
		cmd.Stdin = opts.Stdin
		cmd.Env = opts.Env
		cmd.Dir = opts.Dir
	}

	// Set process group so we can kill child processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}
