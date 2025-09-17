package commander

import (
	"bufio"
	"context"
	"errors"
	"io"
	"syscall"
	"time"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

const defaultStopTimeout = 3 * time.Second

func (p *process) Start() error {
	logrus.Debug("creating process pipes for command")

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		logrus.Debugf("failed to create stdout pipe - error: %v", err)

		return ctxerrors.Wrap(err, "failed to get stdout pipe")
	}

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		logrus.Debugf("failed to create stderr pipe - error: %v", err)

		return ctxerrors.Wrap(err, "failed to get stderr pipe")
	}

	logrus.Debug("starting process")

	if err := p.cmd.Start(); err != nil {
		logrus.Debugf("failed to start process - error: %v", err)

		return ctxerrors.Wrap(err, "failed to start command")
	}

	if p.cmd.Process == nil {
		logrus.Debug("process started but no PID available")
	} else {
		logrus.Debugf("process started successfully - PID: %d", p.cmd.Process.Pid)
	}

	logrus.Debug("starting background goroutines for process")

	go p.readStdout(stdout)
	go p.readStderr(stderr)
	go p.discardInternalOutput()

	logrus.Debug("process initialization complete")

	return nil
}

func (p *process) readStdout(stdout io.ReadCloser) {
	logrus.Debug("starting stdout reader goroutine")

	defer func() {
		logrus.Debug("closing stdout pipe and internal channel")

		_ = stdout.Close()

		close(p.internalStdout)
	}()

	scanner := bufio.NewScanner(stdout)
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		select {
		case <-p.doneCh:
			logrus.Debugf("stdout reader stopping after %d lines (process done)", lineCount)

			return
		case p.internalStdout <- line:
			logrus.Debugf("stdout line %d: %s", lineCount, line)
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Debugf(
			"stdout scanner error after %d lines: %v",
			lineCount,
			err,
		)

		return
	}

	logrus.Debugf(
		"stdout reader finished successfully after %d lines",
		lineCount,
	)
}

// readStderr reads from stderr pipe and sends to internal channel
func (p *process) readStderr(stderr io.ReadCloser) {
	logrus.Debug("starting stderr reader goroutine")

	defer func() {
		logrus.Debug("closing stderr pipe and internal channel")

		_ = stderr.Close() // Ignore close error - nothing we can do

		close(p.internalStderr) // Close internal channel when done
	}()

	scanner := bufio.NewScanner(stderr)
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		select {
		case <-p.doneCh:
			// Process is done, stop reading
			logrus.Debugf(
				"stderr reader stopping after %d lines (process done)",
				lineCount,
			)

			return
		case p.internalStderr <- line:
			// Add to buffer for error reporting
			p.stderrMu.Lock()
			p.stderrBuffer = append(p.stderrBuffer, line)
			p.stderrMu.Unlock()

			logrus.Debugf("stderr line %d: %s", lineCount, line)
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Debugf(
			"stderr scanner error after %d lines: %v",
			lineCount,
			err,
		)

		return
	}

	logrus.Debugf(
		"stderr reader finished successfully after %d lines",
		lineCount,
	)
}

// cleanup performs all resource cleanup operations
func (p *process) cleanup() {
	logrus.Debug("performing resource cleanup")

	// Signal all goroutines to stop and close channels
	logrus.Debug("signaling all goroutines to stop")
	close(p.doneCh)

	// Close stream channels
	p.closeStreamChannels()

	logrus.Debug("cleanup complete")
}

// Stop terminates the process gracefully with context deadline, then kills forcefully
func (p *process) Stop(ctx context.Context) error {
	var stopErr error

	// Use sync.Once to ensure termination happens exactly once
	p.terminateOnce.Do(func() {
		defer p.cleanup()

		stopErr = p.performGracefulStop(ctx)
	})

	return stopErr
}

// performGracefulStop handles the main stop logic
func (p *process) performGracefulStop(ctx context.Context) error {
	logrus.Debug("performing graceful stop")

	if p.cmd.Process == nil {
		logrus.Debug("stop requested but process has no PID - cleaning up anyway")

		return nil
	}

	pid := p.cmd.Process.Pid
	logrus.Debugf("stopping process PID %d", pid)

	timeoutCtx, cancel := p.setupTimeoutContext(ctx)
	if cancel != nil {
		defer cancel()
	}

	if err := p.sendSIGTERM(); err != nil {
		return err
	}

	return p.waitForProcessExit(timeoutCtx)
}

// setupTimeoutContext creates timeout context if none provided
func (p *process) setupTimeoutContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	_, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return context.WithTimeout(ctx, defaultStopTimeout)
	}

	return ctx, nil
}

// sendSIGTERM sends SIGTERM signal to process group
func (p *process) sendSIGTERM() error {
	pid := p.cmd.Process.Pid
	logrus.Debugf("sending SIGTERM to process group PID %d", pid)

	// Kill entire process group to catch child processes
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		logrus.Debugf(
			"failed to send SIGTERM to process group PID %d: %v", pid, err)

		return ctxerrors.Wrap(err, "failed to send SIGTERM")
	}

	return nil
}

// waitForProcessExit waits for process to exit or forces kill on timeout
func (p *process) waitForProcessExit(timeoutCtx context.Context) error {
	pid := p.cmd.Process.Pid
	logrus.Debugf("SIGTERM sent to process PID %d, waiting for graceful shutdown", pid)

	done := make(chan error, 1)

	go func() {
		done <- p.cmdWait()
	}()

	select {
	case err := <-done:
		return p.handleProcessExitResult(err)
	case <-timeoutCtx.Done():
		return p.handleProcessTimeout(timeoutCtx)
	}
}

// handleProcessExitResult processes the result from process exit
func (p *process) handleProcessExitResult(err error) error {
	if isHarmlessWaitError(err) {
		logrus.Debug("process exited cleanly")

		return nil
	}

	if getTerminationSignal(err) == syscall.SIGTERM {
		logrus.Debug("process gracefully terminated by SIGTERM")

		return commonerrors.ErrTerminated
	}

	if isKilledBySignal(err) {
		logrus.Debug("process was killed by SIGKILL")

		return commonerrors.ErrKilled
	}

	logrus.Debugf("process exited with error: %v", err)

	return err
}

// handleProcessTimeout handles timeout or cancellation scenarios
func (p *process) handleProcessTimeout(timeoutCtx context.Context) error {
	pid := p.cmd.Process.Pid

	if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
		logrus.Debugf(
			"graceful shutdown timeout for process PID %d, force killing",
			pid,
		)

		p.forceKillProcess()

		return commonerrors.ErrKilled
	}

	logrus.Debugf(
		"context cancelled for process PID %d, force killing",
		pid,
	)

	p.forceKillProcess()

	return commonerrors.ErrKilled
}

// forceKillProcess immediately kills process group with SIGKILL
func (p *process) forceKillProcess() {
	pid := p.cmd.Process.Pid
	logrus.Debugf("force killing process group PID %d (SIGKILL)", pid)

	// Kill entire process group to catch child processes
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			logrus.Debugf("process group PID %d was already finished", pid)

			return
		}

		logrus.Debugf("failed to force kill process group PID %d: %v", pid, err)

		return
	}

	logrus.Debugf("SIGKILL sent to process group PID %d, waiting for process to exit", pid)

	// Wait for process to exit after SIGKILL
	err := p.cmdWait()
	if err == nil {
		return
	}

	if isKilledBySignal(err) || isTerminatedBySignal(err) || isHarmlessWaitError(err) {
		logrus.Debug("process successfully killed")

		return
	}

	logrus.Debugf("process failed after SIGKILL: %v", err)
}
