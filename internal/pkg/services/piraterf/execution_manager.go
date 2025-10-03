package piraterf

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	dabluveees "github.com/psyb0t/aichteeteapee/server/dabluvee-es"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	_ "github.com/psyb0t/aichteeteapee/server/dabluvee-es/wsunixbridge"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
)

type executionState int32

const (
	executionStateIdle executionState = iota
	executionStateExecuting
	executionStateStopping
)

const (
	stopTimeout             = 3 * time.Second
	stdoutChannelBufferSize = 50
	stderrChannelBufferSize = 10
)

type executionManager struct {
	rpitx            *gorpitx.RPITX
	hub              wshub.Hub
	state            atomic.Int32
	initiatingClient atomic.Value // stores uuid.UUID
	stopRequested    atomic.Bool  // tracks if stop was requested
	outputChannels   struct {
		stdout chan string
		stderr chan string
	}
	streamingCancel context.CancelFunc
	mu              sync.RWMutex
}

func newExecutionManager(
	rpitx *gorpitx.RPITX,
	hub wshub.Hub,
) *executionManager {
	return &executionManager{
		rpitx: rpitx,
		hub:   hub,
	}
}

func (em *executionManager) startExecution(
	ctx context.Context,
	moduleName gorpitx.ModuleName,
	args json.RawMessage,
	timeout int,
	client *wshub.Client,
	callback func() error,
) error {
	// Atomic state transition - only allow if idle
	if !em.state.CompareAndSwap(
		int32(executionStateIdle),
		int32(executionStateExecuting),
	) {
		currentState := executionState(em.state.Load())
		switch currentState {
		case executionStateIdle:
			// This shouldn't happen due to CompareAndSwap, but handle it
			return nil
		case executionStateExecuting:
			em.sendErrorEvent(
				"already executing",
				"execution already in progress",
			)
		case executionStateStopping:
			em.sendErrorEvent(
				"currently stopping",
				"execution currently stopping",
			)
		}

		return nil // Don't return error - just broadcast
	}

	// Validate timeout
	validTimeout := em.validateTimeout(timeout)

	// Store initiating client
	em.initiatingClient.Store(client.ID())

	// Start execution in goroutine
	go em.executeModule(ctx, moduleName, args, validTimeout, client, callback)

	return nil
}

func (em *executionManager) stopExecution(_ *wshub.Client) error {
	currentState := executionState(em.state.Load())

	// Idempotent - return success for already stopped or stopping
	if currentState == executionStateIdle ||
		currentState == executionStateStopping {
		return nil
	}

	// Set stopping state and mark that stop was requested
	em.setState(executionStateStopping)
	em.stopRequested.Store(true)

	// Stop RPITX execution - wait for it to complete
	stopCtx, cancel := context.WithTimeout(
		context.Background(),
		stopTimeout,
	)
	defer cancel()

	if err := em.rpitx.Stop(
		stopCtx,
	); err != nil {
		logrus.WithError(err).
			Error("failed to stop RPITX execution")
	}

	// Reset to idle state - don't send stopped event here since
	// executeModule will handle it
	em.setState(executionStateIdle)

	return nil
}

func (em *executionManager) executeModule(
	ctx context.Context,
	moduleName gorpitx.ModuleName,
	args json.RawMessage,
	timeout time.Duration,
	client *wshub.Client,
	callback func() error,
) {
	defer em.cleanupAfterExecution(client, callback)

	em.logExecutionStart(moduleName, timeout, client)
	em.sendStartedEvent(moduleName, args, client.ID())
	em.setupOutputChannels(ctx)

	err := em.runExecution(ctx, moduleName, args, timeout)
	em.handleExecutionResult(err, client)
}

func (em *executionManager) cleanupAfterExecution(
	client *wshub.Client,
	callback func() error,
) {
	logrus.WithField("clientID", client.ID()).
		Debug("executeModule finished, setting state to idle")

	em.setState(executionStateIdle)
	em.stopRequested.Store(false)

	if callback != nil {
		if err := callback(); err != nil {
			logrus.WithError(err).Error("callback failed")
		}
	}
}

func (em *executionManager) logExecutionStart(
	moduleName gorpitx.ModuleName,
	timeout time.Duration,
	client *wshub.Client,
) {
	logrus.WithFields(logrus.Fields{
		"moduleName": moduleName,
		"timeout":    timeout,
		"clientID":   client.ID(),
	}).Debug("starting executeModule")
}

func (em *executionManager) runExecution(
	ctx context.Context,
	moduleName gorpitx.ModuleName,
	args json.RawMessage,
	timeout time.Duration,
) error {
	execDone := make(chan error, 1)

	em.mu.RLock()
	stdoutCh := em.outputChannels.stdout
	stderrCh := em.outputChannels.stderr
	em.mu.RUnlock()

	em.rpitx.StreamOutputsAsync(stdoutCh, stderrCh)

	go func() {
		logrus.Debug("calling rpitx.Exec")

		err := em.rpitx.Exec(ctx, moduleName, args, timeout)
		logrus.WithError(err).Debug("rpitx.Exec completed")

		execDone <- err
	}()

	return <-execDone
}

func (em *executionManager) handleExecutionResult(
	err error,
	client *wshub.Client,
) {
	logrus.WithFields(logrus.Fields{
		"error":    err,
		"clientID": client.ID(),
	}).Debug("execution completed, processing cleanup")

	em.stopStreaming()

	if err != nil {
		if em.isExpectedTermination(err) {
			logrus.WithError(err).Debug("execution completed with expected termination")
			em.sendStoppedEvent(client.ID())

			return
		}

		logrus.WithError(err).Debug("sending error event")
		em.sendErrorEvent("execution failed", err.Error())

		return
	}

	logrus.WithField("clientID", client.ID()).Debug("sending stopped event")
	em.sendStoppedEvent(client.ID())
}

func (em *executionManager) setupOutputChannels(ctx context.Context) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Create buffered channels
	em.outputChannels.stdout = make(chan string, stdoutChannelBufferSize)
	em.outputChannels.stderr = make(chan string, stderrChannelBufferSize)

	// Create cancellable context for streaming
	streamingCtx, streamingCancel := context.WithCancel(ctx)
	em.streamingCancel = streamingCancel

	// Start output broadcasting goroutine
	go em.streamOutput(streamingCtx)
}

func (em *executionManager) streamOutput(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !em.processOutputChannels(ctx) {
				return
			}
		}

		time.Sleep(time.Millisecond)
	}
}

func (em *executionManager) processOutputChannels(ctx context.Context) bool {
	// Get channel references safely
	em.mu.RLock()
	stdoutCh := em.outputChannels.stdout
	stderrCh := em.outputChannels.stderr
	em.mu.RUnlock()

	if stdoutCh == nil && stderrCh == nil {
		return false
	}

	// Now read from channels safely
	select {
	case <-ctx.Done():
		return false
	case line, ok := <-stdoutCh:
		if !ok {
			return true
		}

		em.sendOutputEvent("stdout", line)
	case line, ok := <-stderrCh:
		if !ok {
			return true
		}

		em.sendOutputEvent("stderr", line)
	default:
		// Avoid busy loop
		return true
	}

	return true
}

func (em *executionManager) stopStreaming() {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Cancel streaming context to stop the streaming goroutine
	if em.streamingCancel != nil {
		em.streamingCancel()
	}

	// Close channels safely - they might already be closed by commander process
	if em.outputChannels.stdout != nil {
		func() {
			defer func() {
				_ = recover() // Channel already closed, ignore
			}()

			close(em.outputChannels.stdout)
		}()

		em.outputChannels.stdout = nil
	}

	if em.outputChannels.stderr != nil {
		func() {
			defer func() {
				_ = recover() // Channel already closed, ignore
			}()

			close(em.outputChannels.stderr)
		}()

		em.outputChannels.stderr = nil
	}
}

func (em *executionManager) isExpectedTermination(err error) bool {
	// If stop was requested, then termination/kill is expected
	if em.stopRequested.Load() {
		return errors.Is(err, commonerrors.ErrTerminated) ||
			errors.Is(err, commonerrors.ErrKilled)
	}

	// Natural timeout is also expected (not an error)
	if errors.Is(err, commonerrors.ErrTimeout) {
		return true
	}

	// Process cleanup errors after timeout are also expected
	if err != nil && (strings.Contains(err.Error(), "process wait failed: exit status") ||
		strings.Contains(err.Error(), "process failed after timeout stop")) {
		return true
	}

	return false
}

func (em *executionManager) validateTimeout(timeout int) time.Duration {
	// Allow 0 for no timeout
	if timeout == 0 {
		return 0
	}

	// Allow any positive timeout value
	return time.Duration(timeout) * time.Second
}

func (em *executionManager) setState(state executionState) {
	em.state.Store(int32(state))
}

// Event broadcasting methods

func (em *executionManager) sendStartedEvent(
	moduleName gorpitx.ModuleName,
	args json.RawMessage,
	clientID uuid.UUID,
) {
	em.hub.BroadcastToAll(dabluveees.NewEvent(
		eventTypeRPITXExecutionStarted,
		rpitxExecutionStartedMessageData{
			ModuleName:         moduleName,
			Args:               args,
			InitiatingClientID: clientID.String(),
			Timestamp:          time.Now().Unix(),
		},
	))
}

func (em *executionManager) sendStoppedEvent(stoppingClientID uuid.UUID) {
	initiatingClientID := uuid.UUID{}

	if val := em.initiatingClient.Load(); val != nil {
		if clientID, ok := val.(uuid.UUID); ok {
			initiatingClientID = clientID
		}
	}

	em.hub.BroadcastToAll(dabluveees.NewEvent(
		eventTypeRPITXExecutionStopped,
		rpitxExecutionStoppedMessageData{
			InitiatingClientID: initiatingClientID.String(),
			StoppingClientID:   stoppingClientID.String(),
			Timestamp:          time.Now().Unix(),
		},
	))
}

// SendError broadcasts an error event to all connected clients.
func (em *executionManager) SendError(errorType, message string) {
	em.sendErrorEvent(errorType, message)
}

func (em *executionManager) sendErrorEvent(errorType, message string) {
	// Log the error details for better visibility
	logrus.WithFields(logrus.Fields{
		"errorType": errorType,
		"message":   message,
	}).Error("RPITX execution error occurred")

	em.hub.BroadcastToAll(dabluveees.NewEvent(
		eventTypeRPITXExecutionError,
		rpitxExecutionErrorMessageData{
			Error:     errorType,
			Message:   message,
			Timestamp: time.Now().Unix(),
		},
	))
}

func (em *executionManager) sendOutputEvent(outputType, line string) {
	em.hub.BroadcastToAll(dabluveees.NewEvent(
		eventTypeRPITXExecutionOutputLine,
		rpitxExecutionOutputLineMessageData{
			Type:      outputType,
			Line:      line,
			Timestamp: time.Now().Unix(),
		},
	))
}
