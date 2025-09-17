package servicemanager

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	errTestService     = errors.New("test error")
	errTestServiceStop = errors.New("stop error")
)

func TestGetInstance(t *testing.T) {
	// Reset singleton before test
	ResetInstance()

	sm1 := GetInstance()
	assert.NotNil(t, sm1)
	assert.NotNil(t, sm1.services)
	assert.NotNil(t, sm1.doneCh)
	assert.Equal(t, 0, len(sm1.services))

	// Test that subsequent calls return the same instance
	sm2 := GetInstance()
	assert.Same(t, sm1, sm2)

	// Test that NewServiceManager also returns the same instance
	sm3 := GetInstance()
	assert.Same(t, sm1, sm3)
}

func TestServiceManager_Add(t *testing.T) {
	tests := []struct {
		name     string
		services []Service
		expected int
	}{
		{
			name:     "add single service",
			services: []Service{NewMockService("test1")},
			expected: 1,
		},
		{
			name:     "add multiple services",
			services: []Service{NewMockService("test1"), NewMockService("test2")},
			expected: 2,
		},
		{
			name:     "add services with duplicate names overwrites",
			services: []Service{NewMockService("test1"), NewMockService("test1")},
			expected: 1,
		},
		{
			name:     "add no services",
			services: []Service{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetInstance()

			sm := GetInstance()
			sm.Add(tt.services...)

			assert.Equal(t, tt.expected, len(sm.services))

			// Verify all services are accessible by name
			for _, svc := range tt.services {
				storedSvc, exists := sm.services[svc.Name()]
				if tt.expected > 0 {
					assert.True(t, exists)
					assert.Equal(t, svc.Name(), storedSvc.Name())
				}
			}
		})
	}
}

func TestServiceManager_Run(t *testing.T) {
	tests := []struct {
		name         string
		services     []Service
		contextSetup func() (context.Context, context.CancelFunc)
		expectError  bool
		stopMethod   string // "context", "stop_method", "service_error"
	}{
		{
			name:     "run single service with context cancellation",
			services: []Service{NewMockService("test1")},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()

				return ctx, cancel
			},
			expectError: false,
			stopMethod:  "context",
		},
		{
			name:     "run multiple services with context cancellation",
			services: []Service{NewMockService("test1"), NewMockService("test2")},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()

				return ctx, cancel
			},
			expectError: false,
			stopMethod:  "context",
		},
		{
			name:     "run service with error",
			services: []Service{NewMockService("failing").WithRunError(errTestService)},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectError: true,
			stopMethod:  "service_error",
		},
		{
			name:     "run with stop method",
			services: []Service{NewMockService("test1")},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectError: false,
			stopMethod:  "stop_method",
		},
		{
			name:     "run with no services",
			services: []Service{},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()

				return ctx, cancel
			},
			expectError: true, // Now expects error since no services means "no enabled services"
			stopMethod:  "context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetInstance()

			sm := GetInstance()
			sm.Add(tt.services...)

			ctx, cancel := tt.contextSetup()
			defer cancel()

			runDone := make(chan error, 1)

			go func() {
				runDone <- sm.Run(ctx)
			}()

			// Give services time to start
			time.Sleep(5 * time.Millisecond)

			// Verify services started (for non-error cases)
			if !tt.expectError && len(tt.services) > 0 {
				for _, svc := range tt.services {
					if mockSvc, ok := svc.(*MockService); ok {
						assert.True(t, mockSvc.WasRunCalled(), "Service %s should have Run called", mockSvc.name)
					}
				}
			}

			// Execute stop method
			switch tt.stopMethod {
			case "stop_method":
				sm.Stop(ctx)
			case "context":
				// Already set up in contextSetup
			case "service_error":
				// Service will error naturally
			}

			// Wait for Run to complete
			select {
			case err := <-runDone:
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			case <-time.After(time.Second):
				t.Fatal("ServiceManager.Run() did not complete within timeout")
			}
		})
	}
}

func TestServiceManager_Stop(t *testing.T) {
	tests := []struct {
		name             string
		services         []Service
		stopTwice        bool
		expectStopErrors bool
	}{
		{
			name:      "stop single service",
			services:  []Service{NewMockService("test1")},
			stopTwice: false,
		},
		{
			name:      "stop multiple services",
			services:  []Service{NewMockService("test1"), NewMockService("test2")},
			stopTwice: false,
		},
		{
			name:      "stop same manager twice (sync.Once behavior)",
			services:  []Service{NewMockService("test1")},
			stopTwice: true,
		},
		{
			name: "stop service with error",
			services: []Service{NewMockService("failing").WithStopError(errTestServiceStop)},
			expectStopErrors: true,
		},
		{
			name:     "stop with no services",
			services: []Service{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetInstance()

			sm := GetInstance()
			sm.Add(tt.services...)

			ctx := context.Background()

			// First run the services if there are any (so they get added to runningServices)
			if len(tt.services) > 0 {
				runCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
				defer cancel()

				// Run services in background so they can be stopped
				runDone := make(chan error, 1)

				go func() {
					runDone <- sm.Run(runCtx)
				}()

				// Give services time to start
				time.Sleep(5 * time.Millisecond)

				// Now stop them
				sm.Stop(ctx)

				// Wait for run to complete
				<-runDone
			} else {
				// For empty services test, just call stop
				sm.Stop(ctx)
			}

			// Verify all services had Stop called (only for non-empty services)
			if len(tt.services) > 0 {
				for _, svc := range tt.services {
					if mockSvc, ok := svc.(*MockService); ok {
						assert.True(t, mockSvc.WasStopCalled(), "Service %s should have Stop called", mockSvc.name)
					}
				}
			}

			// Second stop (if testing sync.Once)
			if tt.stopTwice {
				// Reset the stop called flag for testing
				for _, svc := range tt.services {
					if mockSvc, ok := svc.(*MockService); ok {
						atomic.StoreInt32(&mockSvc.stopCalled, 0)
					}
				}

				// Call stop again - this should be a no-op due to sync.Once
				sm.Stop(ctx)

				// Services should NOT be stopped again due to sync.Once
				for _, svc := range tt.services {
					if mockSvc, ok := svc.(*MockService); ok {
						assert.False(t, mockSvc.WasStopCalled(), "Service %s should NOT have Stop called again", mockSvc.name)
					}
				}
			}
		})
	}
}

func TestServiceManager_Concurrency(t *testing.T) {
	t.Run("concurrent add and run", func(t *testing.T) {
		ResetInstance()

		sm := GetInstance()

		// Add services concurrently
		done := make(chan bool, 10)

		for i := range 10 {
			go func(id int) {
				svc := NewMockService(fmt.Sprintf("service%d", id))
				sm.Add(svc)

				done <- true
			}(i)
		}

		// Wait for all adds to complete
		for range 10 {
			<-done
		}

		// Verify all services were added
		assert.Equal(t, 10, len(sm.services))

		// Run the manager
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := sm.Run(ctx)
		assert.NoError(t, err)
	})
}
