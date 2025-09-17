package app

import (
	"context"
	"sync"

	"github.com/psyb0t/common-go/env"
	"github.com/psyb0t/ctxerrors"
	servicemanager "github.com/psyb0t/piraterf/internal/pkg/service-manager"
	"github.com/psyb0t/piraterf/internal/pkg/services"
	"github.com/sirupsen/logrus"
)

var (
	instance *App      //nolint:gochecknoglobals
	once     sync.Once //nolint:gochecknoglobals
)

type App struct {
	config         config
	wg             sync.WaitGroup
	doneCh         chan struct{}
	stopOnce       sync.Once
	serviceManager *servicemanager.ServiceManager
}

func GetInstance() *App {
	once.Do(func() {
		instance = newApp()
	})

	return instance
}

func newApp() *App {
	// Initialize services first
	services.Init()

	app := &App{
		doneCh:         make(chan struct{}),
		serviceManager: servicemanager.GetInstance(),
	}

	var err error

	app.config, err = parseConfig()
	if err != nil {
		logrus.Fatalf("failed to parse app config: %v", err)
	}

	return app
}

// resetInstance resets the singleton instance for testing purposes.
func resetInstance() {
	once = sync.Once{}
	instance = nil
	// Also reset ServiceManager singleton
	servicemanager.ResetInstance()
}

func (a *App) Run(ctx context.Context) error {
	logrus.Infof("running app with ENV=%s", env.Get())

	defer func() {
		if err := a.Stop(ctx); err != nil {
			logrus.Errorf("failed to stop app: %v", err)
		}
	}()
	defer a.wg.Wait()

	errCh := make(chan error, 1)
	defer close(errCh)

	a.wg.Add(1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer a.wg.Done()

		if err := a.serviceManager.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return ctxerrors.Wrap(err, "failed to run app")
	case <-a.doneCh:
		return nil
	}
}

func (a *App) Stop(ctx context.Context) error {
	a.stopOnce.Do(func() {
		logrus.Info("stopping app")
		defer logrus.Info("stopped app")

		close(a.doneCh)
		a.serviceManager.Stop(ctx)
		a.wg.Wait()
	})

	return nil
}
