package piraterf

import (
	"context"
	"testing"
	"time"

	"github.com/psyb0t/aichteeteapee/server/websocket"
	"github.com/psyb0t/common-go/env"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/gorpitx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
