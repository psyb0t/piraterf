package piraterf

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/common-go/env"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Suppress debug logs in tests
	logrus.SetLevel(logrus.WarnLevel)
}

func TestExecutionManager_StopStreaming_DoubleClose(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Create a mock hub
	hub := websocket.NewHub("test")

	// Create ExecutionManager with mock RPITX
	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Setup output channels to simulate the scenario
	ctx := context.Background()
	em.setupOutputChannels(ctx)

	// Simulate channels being closed by commander process
	// (this would normally happen when the process finishes)
	close(em.outputChannels.stdout)
	close(em.outputChannels.stderr)

	// Now call stopStreaming - this should NOT panic even though channels are already closed
	require.NotPanics(t, func() {
		em.stopStreaming()
	}, "stopStreaming should not panic when channels are already closed")

	// Verify channels are set to nil after cleanup
	em.mu.RLock()
	assert.Nil(t, em.outputChannels.stdout, "stdout channel should be nil after cleanup")
	assert.Nil(t, em.outputChannels.stderr, "stderr channel should be nil after cleanup")
	em.mu.RUnlock()
}

func TestExecutionManager_StopStreaming_NormalClose(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Create a mock hub
	hub := websocket.NewHub("test")

	// Create ExecutionManager with mock RPITX
	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Setup output channels
	ctx := context.Background()
	em.setupOutputChannels(ctx)

	// Call stopStreaming - this should work normally
	require.NotPanics(t, func() {
		em.stopStreaming()
	}, "stopStreaming should not panic during normal operation")

	// Verify channels are set to nil after cleanup
	em.mu.RLock()
	assert.Nil(t, em.outputChannels.stdout, "stdout channel should be nil after cleanup")
	assert.Nil(t, em.outputChannels.stderr, "stderr channel should be nil after cleanup")
	em.mu.RUnlock()
}

func TestExecutionManager_StopStreaming_MultipleCallsSafe(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Create a mock hub
	hub := websocket.NewHub("test")

	// Create ExecutionManager with mock RPITX
	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Setup output channels
	ctx := context.Background()
	em.setupOutputChannels(ctx)

	// Call stopStreaming multiple times - should be safe
	require.NotPanics(t, func() {
		em.stopStreaming()
		em.stopStreaming() // Second call should be safe
		em.stopStreaming() // Third call should be safe
	}, "multiple calls to stopStreaming should be safe")
}

func TestExecutionManager_StreamOutput_ContextCancellation(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	// Create a mock hub
	hub := websocket.NewHub("test")

	// Create ExecutionManager with mock RPITX
	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Setup channels manually
	em.mu.Lock()
	em.outputChannels.stdout = make(chan string, 1)
	em.outputChannels.stderr = make(chan string, 1)
	em.mu.Unlock()

	// Start streaming in goroutine
	done := make(chan struct{})

	go func() {
		defer close(done)

		em.streamOutput(ctx)
	}()

	// Cancel context to stop streaming
	cancel()

	// Wait for streaming to stop with timeout
	select {
	case <-done:
		// Success - streaming stopped gracefully
	case <-time.After(1 * time.Second):
		t.Fatal("streamOutput did not stop after context cancellation")
	}

	// Clean up channels
	em.mu.Lock()

	if em.outputChannels.stdout != nil {
		close(em.outputChannels.stdout)
	}

	if em.outputChannels.stderr != nil {
		close(em.outputChannels.stderr)
	}

	em.mu.Unlock()
}

func TestExecutionManager_IsExpectedTermination(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	tests := []struct {
		name          string
		err           error
		stopRequested bool
		expected      bool
	}{
		{
			name:          "timeout error is always expected",
			err:           commonerrors.ErrTimeout,
			stopRequested: false,
			expected:      true,
		},
		{
			name:          "terminated with stop requested is expected",
			err:           commonerrors.ErrTerminated,
			stopRequested: true,
			expected:      true,
		},
		{
			name:          "killed with stop requested is expected",
			err:           commonerrors.ErrKilled,
			stopRequested: true,
			expected:      true,
		},
		{
			name:          "terminated without stop requested is unexpected",
			err:           commonerrors.ErrTerminated,
			stopRequested: false,
			expected:      false,
		},
		{
			name:          "killed without stop requested is unexpected",
			err:           commonerrors.ErrKilled,
			stopRequested: false,
			expected:      false,
		},
		{
			name:          "other errors are unexpected",
			err:           assert.AnError,
			stopRequested: false,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em.stopRequested.Store(tt.stopRequested)
			result := em.isExpectedTermination(tt.err)
			assert.Equal(t, tt.expected, result, "isExpectedTermination should return %v for %s", tt.expected, tt.name)
		})
	}
}

func TestExecutionManager_StartExecution(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name          string
		initialState  executionState
		moduleName    gorpitx.ModuleName
		args          json.RawMessage
		timeout       int
		expectError   bool
		expectedState executionState
	}{
		{
			name:          "start with invalid args returns to idle",
			initialState:  executionStateIdle,
			moduleName:    gorpitx.ModuleNamePIFMRDS,
			args:          json.RawMessage(`{"freq": 88.0}`), // Missing required audio field
			timeout:       10,
			expectError:   false,
			expectedState: executionStateIdle, // Returns to idle after validation failure
		},
		{
			name:          "start while already executing",
			initialState:  executionStateExecuting,
			moduleName:    gorpitx.ModuleNamePIFMRDS,
			args:          json.RawMessage(`{"freq": 88.0}`),
			timeout:       10,
			expectError:   false, // Function returns nil but broadcasts error
			expectedState: executionStateExecuting, // State remains unchanged
		},
		{
			name:          "start while stopping",
			initialState:  executionStateStopping,
			moduleName:    gorpitx.ModuleNamePIFMRDS,
			args:          json.RawMessage(`{"freq": 88.0}`),
			timeout:       10,
			expectError:   false, // Function returns nil but broadcasts error
			expectedState: executionStateStopping, // State remains unchanged
		},
		{
			name:          "start with invalid file returns to idle",
			initialState:  executionStateIdle,
			moduleName:    gorpitx.ModuleNameSPECTRUMPAINT,
			args:          json.RawMessage(`{"pictureFile": "/path/to/image.png"}`), // File doesn't exist
			timeout:       0,
			expectError:   false,
			expectedState: executionStateIdle, // Returns to idle after validation failure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := websocket.NewHub("test")
			rpitx := gorpitx.GetInstance()
			em := newExecutionManager(rpitx, hub)

			// Set initial state
			em.setState(tt.initialState)

			// Create a mock client
			client := &websocket.Client{}

			// Call startExecution
			err := em.startExecution(
				context.Background(),
				tt.moduleName,
				tt.args,
				tt.timeout,
				client,
				nil, // callback
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Give goroutine a moment to start - execution validation might take time
			time.Sleep(50 * time.Millisecond)

			// Check final state
			finalState := executionState(em.state.Load())
			if tt.expectedState == executionStateExecuting && finalState != executionStateExecuting {
				// If we expected executing but didn't get it, wait a bit longer for async operations
				time.Sleep(100 * time.Millisecond)
				finalState = executionState(em.state.Load())
			}
			assert.Equal(t, tt.expectedState, finalState, "Final state should match expected")

			// Cleanup - stop any started execution
			if finalState == executionStateExecuting {
				em.stopExecution(client)
				// Wait for stop to complete
				time.Sleep(100 * time.Millisecond)
			}

			hub.Close()
		})
	}
}

func TestExecutionManager_StopExecution(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	tests := []struct {
		name          string
		initialState  executionState
		expectError   bool
		expectedState executionState
	}{
		{
			name:          "stop when idle (idempotent)",
			initialState:  executionStateIdle,
			expectError:   false,
			expectedState: executionStateIdle,
		},
		{
			name:          "stop when executing",
			initialState:  executionStateExecuting,
			expectError:   false,
			expectedState: executionStateIdle,
		},
		{
			name:          "stop when already stopping (idempotent)",
			initialState:  executionStateStopping,
			expectError:   false,
			expectedState: executionStateStopping,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := websocket.NewHub("test")
			rpitx := gorpitx.GetInstance()
			em := newExecutionManager(rpitx, hub)

			// Set initial state
			em.setState(tt.initialState)

			// Create a mock client
			client := &websocket.Client{}

			// Call stopExecution
			err := em.stopExecution(client)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check final state
			finalState := executionState(em.state.Load())
			assert.Equal(t, tt.expectedState, finalState, "Final state should match expected")

			hub.Close()
		})
	}
}

func TestExecutionManager_ValidateTimeout(t *testing.T) {
	// Set ENV=dev to avoid root check
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "zero timeout (no timeout)",
			timeout:  0,
			expected: 0,
		},
		{
			name:     "positive timeout",
			timeout:  30,
			expected: 30 * time.Second,
		},
		{
			name:     "small positive timeout",
			timeout:  1,
			expected: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := em.validateTimeout(tt.timeout)
			assert.Equal(t, tt.expected, result, "validateTimeout should return correct duration")
		})
	}

	hub.Close()
}

func TestSendStoppedEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	clientID := uuid.New()

	require.NotPanics(t, func() {
		em.sendStoppedEvent(clientID)
	})
}

func TestSendError(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Test that SendError doesn't panic
	assert.NotPanics(t, func() {
		em.SendError("TEST_ERROR", "test error message")
	})
}

func TestSendOutputEvent(t *testing.T) {
	t.Setenv(env.EnvVarName, env.EnvTypeDev)

	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Test that sendOutputEvent doesn't panic
	assert.NotPanics(t, func() {
		em.sendOutputEvent("stdout", "test output line")
	})
}

func TestExecutionManagerStreamOutput(t *testing.T) {
	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	// Test streamOutput with context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start streamOutput in goroutine
	done := make(chan bool)
	go func() {
		em.streamOutput(ctx)
		done <- true
	}()

	// Cancel context to stop streaming
	cancel()

	// Wait for function to return
	select {
	case <-done:
		// Success - function returned when context was cancelled
	case <-time.After(1 * time.Second):
		t.Fatal("streamOutput did not return when context was cancelled")
	}
}

func TestExecutionManagerSendStoppedEvent(t *testing.T) {
	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	stoppingClientID := uuid.New()

	// Test with no initiating client
	require.NotPanics(t, func() {
		em.sendStoppedEvent(stoppingClientID)
	})

	// Test with initiating client set
	initiatingClientID := uuid.New()
	em.initiatingClient.Store(initiatingClientID)

	require.NotPanics(t, func() {
		em.sendStoppedEvent(stoppingClientID)
	})
}

func TestProcessOutputChannels(t *testing.T) {
	hub := websocket.NewHub("test")
	defer hub.Close()

	rpitx := gorpitx.GetInstance()
	em := newExecutionManager(rpitx, hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name           string
		setupChannels  func() (chan string, chan string)
		expectReturn   bool
	}{
		{
			name: "nil channels should return false",
			setupChannels: func() (chan string, chan string) {
				return nil, nil
			},
			expectReturn: false,
		},
		{
			name: "context cancelled should return false",
			setupChannels: func() (chan string, chan string) {
				stdoutCh := make(chan string, 1)
				stderrCh := make(chan string, 1)
				return stdoutCh, stderrCh
			},
			expectReturn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdoutCh, stderrCh := tt.setupChannels()

			// Set up channels in execution manager
			em.mu.Lock()
			em.outputChannels.stdout = stdoutCh
			em.outputChannels.stderr = stderrCh
			em.mu.Unlock()

			if tt.name == "context cancelled should return false" {
				// Cancel context for this test
				testCtx, testCancel := context.WithCancel(context.Background())
				testCancel() // Cancel immediately

				result := em.processOutputChannels(testCtx)
				assert.Equal(t, tt.expectReturn, result)
			} else {
				result := em.processOutputChannels(ctx)
				assert.Equal(t, tt.expectReturn, result)
			}
		})
	}
}
