package servicemanager

import (
	"context"
	"sync/atomic"
	"time"
)

// TestService is a minimal mock for simple testing needs.
type TestService struct{ name string }

func NewTestService(name string) *TestService {
	return &TestService{name: name}
}

func (s *TestService) Name() string { return s.name }
func (s *TestService) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (s *TestService) Stop(_ context.Context) error { return nil }

// MockService is a full-featured mock for comprehensive testing.
type MockService struct {
	name       string
	running    int32
	runCalled  int32
	stopCalled int32
	runError   error
	stopError  error
	stopCh     chan struct{}
	runDelay   time.Duration
}

func NewMockService(name string) *MockService {
	return &MockService{
		name:   name,
		stopCh: make(chan struct{}),
	}
}

func (m *MockService) WithRunError(err error) *MockService {
	m.runError = err

	return m
}

func (m *MockService) WithStopError(err error) *MockService {
	m.stopError = err

	return m
}

func (m *MockService) WithRunDelay(delay time.Duration) *MockService {
	m.runDelay = delay

	return m
}

func (m *MockService) Name() string {
	return m.name
}

func (m *MockService) Run(ctx context.Context) error {
	atomic.StoreInt32(&m.runCalled, 1)
	atomic.StoreInt32(&m.running, 1)

	if m.runError != nil {
		return m.runError
	}

	if m.runDelay > 0 {
		time.Sleep(m.runDelay)
	}

	select {
	case <-ctx.Done():
		return nil
	case <-m.stopCh:
		return nil
	}
}

func (m *MockService) Stop(_ context.Context) error {
	atomic.StoreInt32(&m.stopCalled, 1)
	atomic.StoreInt32(&m.running, 0)

	if m.stopCh != nil {
		close(m.stopCh)
	}

	return m.stopError
}

func (m *MockService) WasRunCalled() bool {
	return atomic.LoadInt32(&m.runCalled) == 1
}

func (m *MockService) WasStopCalled() bool {
	return atomic.LoadInt32(&m.stopCalled) == 1
}

func (m *MockService) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}
