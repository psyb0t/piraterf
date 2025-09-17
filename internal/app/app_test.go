package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	servicemanager "github.com/psyb0t/piraterf/internal/pkg/service-manager"
	"github.com/stretchr/testify/assert"
)

// mockService implements the Service interface for testing.
type mockService struct {
	name     string
	running  int32
	stopCh   chan struct{}
	runError error
}

func newMockService(name string) *mockService {
	return &mockService{
		name:   name,
		stopCh: make(chan struct{}),
	}
}

func (m *mockService) Name() string {
	return m.name
}

func (m *mockService) Run(ctx context.Context) error {
	atomic.StoreInt32(&m.running, 1)

	if m.runError != nil {
		return m.runError
	}

	select {
	case <-ctx.Done():
		return nil
	case <-m.stopCh:
		return nil
	}
}

func (m *mockService) Stop(_ context.Context) error {
	close(m.stopCh)
	atomic.StoreInt32(&m.running, 0)

	return nil
}

// mockServiceWithError implements the Service interface for testing error cases.
type mockServiceWithError struct {
	name     string
	runError error
}

func (m *mockServiceWithError) Name() string {
	return m.name
}

func (m *mockServiceWithError) Run(_ context.Context) error {
	return m.runError
}

func (m *mockServiceWithError) Stop(_ context.Context) error {
	return nil
}

// createTestApp creates an app with mock services instead of real ones.
func createTestApp() *App {
	app := &App{
		doneCh:         make(chan struct{}),
		serviceManager: servicemanager.GetInstance(),
	}

	// Parse empty config
	cfg, _ := parseConfig()
	app.config = cfg

	app.setupTestServices()

	return app
}

func (a *App) setupTestServices() {
	// Use mock services instead of real ones
	mockSvc1 := newMockService("TestService1")
	mockSvc2 := newMockService("TestService2")

	a.serviceManager.Add(mockSvc1, mockSvc2)
}

func TestApp_Run(t *testing.T) {
	// Reset singleton before each test
	tests := []struct {
		name        string
		setupFunc   func() (context.Context, context.CancelFunc)
		runFunc     func(t *testing.T, app *App, ctx context.Context)
		expectError bool
	}{
		{
			name: "context cancellation",
			setupFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately to test graceful shutdown

				return ctx, cancel
			},
			runFunc: func(t *testing.T, app *App, ctx context.Context) {
				t.Helper()
				err := app.Run(ctx)
				assert.NoError(t, err)
			},
			expectError: false,
		},
		{
			name: "stop via done channel",
			setupFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			runFunc: func(t *testing.T, app *App, ctx context.Context) {
				t.Helper()
				// Start app in goroutine
				done := make(chan error, 1)
				go func() {
					done <- app.Run(ctx)
				}()

				// Give app time to start
				time.Sleep(10 * time.Millisecond)

				// Stop the app
				err := app.Stop(ctx)
				assert.NoError(t, err)

				// Wait for run to complete
				select {
				case err := <-done:
					assert.NoError(t, err)
				case <-time.After(time.Second):
					t.Fatal("app.Run() did not complete within timeout")
				}
			},
			expectError: false,
		},
		{
			name: "timeout context",
			setupFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			runFunc: func(t *testing.T, app *App, ctx context.Context) {
				t.Helper()
				err := app.Run(ctx)
				assert.NoError(t, err)
			},
			expectError: false,
		},
		{
			name: "service error propagation",
			setupFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			runFunc: func(t *testing.T, _ *App, ctx context.Context) {
				t.Helper()
				// Create app with failing service
				failingApp := &App{
					doneCh:         make(chan struct{}),
					serviceManager: servicemanager.GetInstance(),
				}
				cfg, _ := parseConfig()
				failingApp.config = cfg

				// Add a service that returns an error
				failingSvc := &mockServiceWithError{name: "failing", runError: assert.AnError}
				failingApp.serviceManager.Add(failingSvc)

				err := failingApp.Run(ctx)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to run app")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetInstance()

			app := createTestApp()

			ctx, cancel := tt.setupFunc()
			defer cancel()

			tt.runFunc(t, app, ctx)
		})
	}
}

func TestApp_Stop(t *testing.T) {
	t.Run("stop once", func(t *testing.T) {
		resetInstance()

		app := createTestApp()

		ctx := context.Background()

		// First stop should work
		err := app.Stop(ctx)
		assert.NoError(t, err)

		// Second stop should also work (sync.Once behavior)
		err = app.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestApp_GetInstance(t *testing.T) {
	// Reset singleton before test
	resetInstance()

	t.Run("successful app instance creation", func(t *testing.T) {
		app := GetInstance()
		assert.NotNil(t, app)
		assert.NotNil(t, app.serviceManager)
		assert.NotNil(t, app.doneCh)

		// Test that subsequent calls return the same instance
		app2 := GetInstance()
		assert.Same(t, app, app2)
	})
}

func TestApp_RunAndStop_Integration(t *testing.T) {
	t.Run("full lifecycle", func(t *testing.T) {
		resetInstance()

		app := createTestApp()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start app in goroutine
		runDone := make(chan error, 1)

		go func() {
			runDone <- app.Run(ctx)
		}()

		// Let app start
		time.Sleep(50 * time.Millisecond)

		// Stop app
		stopErr := app.Stop(ctx)
		assert.NoError(t, stopErr)

		// Wait for run to complete
		select {
		case runErr := <-runDone:
			assert.NoError(t, runErr)
		case <-ctx.Done():
			t.Fatal("test timed out")
		}
	})
}

func TestApp_ServiceActivity(t *testing.T) {
	tests := []struct {
		name         string
		serviceNames []string
		stopMethod   string // "app_stop" or "context_cancel"
		expectError  bool
	}{
		{
			name:         "single service run and stop correctly",
			serviceNames: []string{"TestService1"},
			stopMethod:   "app_stop",
			expectError:  false,
		},
		{
			name:         "multiple services run and stop correctly",
			serviceNames: []string{"TestService1", "TestService2"},
			stopMethod:   "app_stop",
			expectError:  false,
		},
		{
			name:         "services stop via context cancellation",
			serviceNames: []string{"TestService1", "TestService2"},
			stopMethod:   "context_cancel",
			expectError:  false,
		},
		{
			name:         "many services run and stop correctly",
			serviceNames: []string{"Service1", "Service2", "Service3", "Service4"},
			stopMethod:   "app_stop",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and clear services before test
			servicemanager.ResetInstance()
			servicemanager.GetInstance().ClearServices()

			// Create mock services that track their activity
			var mockServices []*mockServiceWithActivity
			for _, name := range tt.serviceNames {
				mockServices = append(mockServices, &mockServiceWithActivity{name: name})
			}

			// Create app manually and add mock services
			app := &App{
				doneCh:         make(chan struct{}),
				serviceManager: servicemanager.GetInstance(),
			}
			cfg, _ := parseConfig()
			app.config = cfg

			// Convert to Service interface for Add method
			var servicesForAdd []servicemanager.Service
			for _, svc := range mockServices {
				servicesForAdd = append(servicesForAdd, svc)
			}

			app.serviceManager.Add(servicesForAdd...)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start app in goroutine
			done := make(chan error, 1)

			go func() {
				done <- app.Run(ctx)
			}()

			// Give services time to start
			time.Sleep(20 * time.Millisecond)

			// Verify all services are running
			for i, mockSvc := range mockServices {
				assert.Equal(t, int32(1), atomic.LoadInt32(&mockSvc.running),
					"Service %d (%s) should be running", i, mockSvc.name)
			}

			// Stop using specified method
			var stopErr error

			switch tt.stopMethod {
			case "app_stop":
				stopErr = app.Stop(ctx)
			case "context_cancel":
				cancel()
			}

			if !tt.expectError {
				assert.NoError(t, stopErr)
			}

			// Wait for run to complete
			select {
			case err := <-done:
				if !tt.expectError {
					assert.NoError(t, err)
				}
			case <-time.After(time.Second):
				t.Fatal("app.Run() did not complete within timeout")
			}

			// Verify all services have stopped
			for i, mockSvc := range mockServices {
				assert.Equal(t, int32(0), atomic.LoadInt32(&mockSvc.running),
					"Service %d (%s) should be stopped", i, mockSvc.name)
			}
		})
	}
}

// mockServiceWithActivity is a mock service that tracks its running state.
type mockServiceWithActivity struct {
	name    string
	running int32
	stopCh  chan struct{}
}

func (m *mockServiceWithActivity) Name() string {
	return m.name
}

func (m *mockServiceWithActivity) Run(ctx context.Context) error {
	m.stopCh = make(chan struct{})
	atomic.StoreInt32(&m.running, 1)

	select {
	case <-ctx.Done():
		return nil
	case <-m.stopCh:
		return nil
	}
}

func (m *mockServiceWithActivity) Stop(_ context.Context) error {
	if m.stopCh != nil {
		close(m.stopCh)
	}

	atomic.StoreInt32(&m.running, 0)

	return nil
}
