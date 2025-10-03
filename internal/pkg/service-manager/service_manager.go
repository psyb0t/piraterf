package servicemanager

import (
	"context"
	"sync"

	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

var (
	serviceManagerInstance *ServiceManager //nolint:gochecknoglobals
	serviceManagerOnce     sync.Once       //nolint:gochecknoglobals
)

type Service interface {
	Name() string
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ServiceManager struct {
	services             map[string]Service
	servicesMutex        sync.RWMutex
	runningServices      []Service
	runningServicesMutex sync.RWMutex
	wg                   sync.WaitGroup
	doneCh               chan struct{}
	stopOnce             sync.Once
}

func GetInstance() *ServiceManager {
	serviceManagerOnce.Do(func() {
		serviceManagerInstance = &ServiceManager{
			services: make(map[string]Service),
			doneCh:   make(chan struct{}),
		}
	})

	return serviceManagerInstance
}

// ResetInstance resets the singleton instance for testing purposes.
func ResetInstance() {
	serviceManagerOnce = sync.Once{}
	serviceManagerInstance = nil
}

// ClearServices clears all services from the service manager for testing.
func (s *ServiceManager) ClearServices() {
	s.servicesMutex.Lock()
	defer s.servicesMutex.Unlock()

	s.services = make(map[string]Service)
}

func (s *ServiceManager) Add(services ...Service) {
	s.servicesMutex.Lock()
	defer s.servicesMutex.Unlock()

	for _, service := range services {
		s.services[service.Name()] = service
	}
}

func (s *ServiceManager) Run(ctx context.Context) error {
	logrus.Info("running services")

	s.servicesMutex.RLock()
	defer s.servicesMutex.RUnlock()

	errCh := make(chan error, 1)
	defer close(errCh)

	defer s.wg.Wait()
	defer s.Stop(ctx)

	services := make([]Service, 0, len(s.services))
	for _, service := range s.services {
		services = append(services, service)
	}

	if len(services) == 0 {
		return ErrNoEnabledServices
	}

	s.runServices(ctx, services, errCh)

	select {
	case <-ctx.Done():
		logrus.Info("services run context done")

		return nil
	case err := <-errCh:
		return ctxerrors.Wrap(err, "service failed")
	case <-s.doneCh:
		return nil
	}
}

func (s *ServiceManager) runServices(
	ctx context.Context,
	services []Service,
	errCh chan<- error,
) {
	s.runningServicesMutex.Lock()
	defer s.runningServicesMutex.Unlock()

	for _, service := range services {
		s.wg.Add(1)

		go func(service Service) {
			defer s.wg.Done()

			if err := service.Run(ctx); err != nil {
				errCh <- err
			}
		}(service)

		s.runningServices = append(s.runningServices, service)
	}
}

func (s *ServiceManager) Stop(ctx context.Context) {
	s.stopOnce.Do(func() {
		logrus.Info("stopping services")
		defer logrus.Info("stopped services")

		close(s.doneCh)

		s.runningServicesMutex.RLock()
		defer s.runningServicesMutex.RUnlock()

		var wg sync.WaitGroup
		for _, service := range s.runningServices {
			wg.Add(1)

			go func(service Service) {
				defer wg.Done()

				if err := service.Stop(ctx); err != nil {
					logrus.Errorf("failed to stop service %s: %v",
						service.Name(), err)
				}
			}(service)
		}

		wg.Wait()
	})
}
