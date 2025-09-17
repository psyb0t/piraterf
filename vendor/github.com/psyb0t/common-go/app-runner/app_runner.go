package runner

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gonfiguration"
	"github.com/sirupsen/logrus"
)

var ErrShutdownTimeout = errors.New("shutdown timeout")

const (
	envVarNameAppRunnerShutdownTimeout = "APPRUNNER_SHUTDOWNTIMEOUT"
	defaultShutdownTimeout             = 10 * time.Second
)

type Runnable interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

type config struct {
	ShutdownTimeout time.Duration `env:"APPRUNNER_SHUTDOWNTIMEOUT"`
}

func Run(runnable Runnable) error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &appRunner{
		runnable:        runnable,
		shutdownTimeout: cfg.ShutdownTimeout,
	}

	return runner.run(ctx)
}

type appRunner struct {
	runnable        Runnable
	shutdownTimeout time.Duration
}

func getConfig() (*config, error) {
	cfg := &config{}

	gonfiguration.SetDefaults(map[string]any{
		envVarNameAppRunnerShutdownTimeout: defaultShutdownTimeout,
	})

	err := gonfiguration.Parse(cfg)
	if err != nil {
		return nil, ctxerrors.Wrap(err, "could not configparser.Parse")
	}

	return cfg, nil
}

func (r *appRunner) run(ctx context.Context) error {
	sigCh := r.setupSignalHandling()
	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	// Start the application
	wg.Add(1)

	go r.runApp(ctx, &wg, errCh)

	// Wait for shutdown signal or error
	shutdownErr := r.waitForShutdown(sigCh, errCh)

	// Perform graceful shutdown
	return r.performGracefulShutdown(ctx, &wg, shutdownErr)
}

func (r *appRunner) setupSignalHandling() chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return sigCh
}

func (r *appRunner) runApp(ctx context.Context, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()
	defer close(errCh)

	logrus.Info("starting application...")

	errCh <- r.runnable.Run(ctx)
}

func (r *appRunner) waitForShutdown(sigCh chan os.Signal, errCh chan error) error {
	select {
	case sig := <-sigCh:
		logrus.Infof("received signal: %v", sig)

		return nil
	case err := <-errCh:
		if err != nil {
			logrus.Errorf("application encountered an error: %v", err)
		}

		return err
	}
}

func (r *appRunner) performGracefulShutdown(ctx context.Context, wg *sync.WaitGroup, shutdownErr error) error {
	logrus.Info("initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, r.shutdownTimeout)
	defer cancel()

	wg.Add(1)

	errCh := make(chan error, 1)

	go func() {
		defer wg.Done()

		if err := r.runnable.Stop(ctx); err != nil {
			errCh <- err
		}
	}()
	// Call Stop on the runnable

	// Wait for all goroutines to finish
	doneCh := make(chan struct{})

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case err := <-errCh:
		return ctxerrors.Wrap(shutdownErr, err.Error())
	case <-shutdownCtx.Done():
		err := shutdownCtx.Err()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				logrus.Error(ErrShutdownTimeout.Error())

				return ErrShutdownTimeout
			}

			if shutdownErr != nil {
				return ctxerrors.Wrap(shutdownErr, err.Error())
			}

			return ctxerrors.Wrap(err, "shutdown error")
		}

	case <-doneCh:
		logrus.Info("shutdown completed successfully")
	}

	return shutdownErr
}
