package gorpitx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/psyb0t/commander"
	"github.com/psyb0t/common-go/env"
	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

const (
	minFreqKHz            = 5
	maxFreqKHz            = 1500000
	gracefulStopTimeout   = 3 * time.Second
	streamingPollInterval = 10 * time.Millisecond
)

type Module interface {
	ParseArgs(json.RawMessage) ([]string, error)
}

type ModuleName = string

type RPITX struct {
	config      Config
	commander   commander.Commander
	modules     map[ModuleName]Module
	isExecuting atomic.Bool
	process     commander.Process
	processMu   sync.RWMutex
}

func newRPITX() *RPITX {
	config, err := parseConfig()
	if err != nil {
		panic(err)
	}

	// Check if running as root in production
	if !env.IsDev() && os.Geteuid() != 0 {
		panic("PIrateRF must be run as root in production mode")
	}

	return &RPITX{
		config:    config,
		commander: commander.New(),
		modules: map[ModuleName]Module{
			ModuleNamePIFMRDS:       &PIFMRDS{},
			ModuleNameTUNE:          &TUNE{},
			ModuleNameMORSE:         &MORSE{},
			ModuleNameSPECTRUMPAINT: &SPECTRUMPAINT{},
		},
	}
}

var (
	instance *RPITX    //nolint:gochecknoglobals
	once     sync.Once //nolint:gochecknoglobals
)

func GetInstance() *RPITX {
	once.Do(func() {
		instance = newRPITX()
	})

	return instance
}

func (r *RPITX) GetSupportedModules() []ModuleName {
	modules := make([]ModuleName, 0, len(r.modules))
	for name := range r.modules {
		modules = append(modules, name)
	}

	return modules
}

func (r *RPITX) IsSupportedModule(name ModuleName) bool {
	_, exists := r.modules[name]

	return exists
}

func (r *RPITX) Exec(
	ctx context.Context,
	name ModuleName,
	args []byte,
	timeout time.Duration,
) error {
	if !r.isExecuting.CompareAndSwap(false, true) {
		return ErrExecuting
	}

	defer r.cleanupExecution(ctx)

	logrus.Debugf("executing module %s with args %s", name, args)
	defer logrus.Debugf("finished executing module %s", name)

	cmdName, cmdArgs, err := r.prepareCommand(name, args)
	if err != nil {
		return err
	}

	if err := r.startProcess(ctx, cmdName, cmdArgs); err != nil {
		return err
	}

	// Handle timeout manually if specified
	if timeout > 0 {
		return r.waitWithTimeout(ctx, timeout)
	}

	if err := r.process.Wait(); err != nil {
		return ctxerrors.Wrap(err, "failed to wait for process")
	}

	return nil
}

func (r *RPITX) cleanupExecution(ctx context.Context) {
	r.processMu.Lock()

	if r.process != nil {
		// fkin kill the fuckin' process
		if err := r.process.Kill(ctx); err != nil {
			logrus.Errorf("failed to kill the fuckin' process: %v", err)
		}
	}

	r.process = nil
	r.processMu.Unlock()

	r.isExecuting.Store(false)
}

func (r *RPITX) prepareCommand(
	name ModuleName,
	args []byte,
) (string, []string, error) {
	if !r.IsSupportedModule(name) {
		return "", nil, ctxerrors.Wrap(ErrUnknownModule, name)
	}

	module := r.modules[name]

	parsedArgs, err := module.ParseArgs(args)
	if err != nil {
		return "", nil, ctxerrors.Wrap(err, "failed to parse args")
	}

	var (
		cmdName string
		cmdArgs []string
	)

	if env.IsDev() {
		cmdName, cmdArgs = r.getMockExecCmd(name, parsedArgs)

		return cmdName, cmdArgs, nil
	}

	binaryPath := filepath.Join(r.config.Path, name)

	// Wrap with stdbuf for line buffering
	cmdName = "stdbuf"

	cmdArgs = append([]string{"-oL", binaryPath}, parsedArgs...)

	logrus.Debugf("production command prepared: %s %v", cmdName, cmdArgs)

	return cmdName, cmdArgs, nil
}

func (r *RPITX) startProcess(
	ctx context.Context,
	cmdName string,
	cmdArgs []string,
) error {
	r.processMu.Lock()
	process, err := r.commander.Start(
		ctx,
		cmdName,
		cmdArgs,
	)
	r.process = process
	r.processMu.Unlock()

	if err != nil {
		return ctxerrors.Wrap(err, "failed to start process")
	}

	return nil
}

func (r *RPITX) StreamOutputs(stdout, stderr chan<- string) {
	if !r.isExecuting.Load() {
		logrus.WithError(ErrNotExecuting).Warn("not executing")

		return
	}

	r.processMu.RLock()
	process := r.process
	r.processMu.RUnlock()

	if process != nil {
		process.Stream(stdout, stderr)

		return
	}

	logrus.Warn("no process to stream")
}

// StreamOutputsAsync starts streaming outputs for the currently executing process.
// This is a convenience method that can be called before or during execution.
// It will wait for execution to start and then begin streaming.
func (r *RPITX) StreamOutputsAsync(stdout, stderr chan<- string) {
	go func() {
		// Wait for execution to start
		for !r.isExecuting.Load() {
			time.Sleep(streamingPollInterval)
		}

		// Wait a bit more for the process to be created
		for {
			r.processMu.RLock()
			process := r.process
			r.processMu.RUnlock()

			if process != nil {
				process.Stream(stdout, stderr)

				break
			}

			if !r.isExecuting.Load() {
				// Execution finished before we could get the process
				logrus.Warn("execution finished before streaming could start")

				break
			}

			time.Sleep(streamingPollInterval)
		}
	}()
}

func (r *RPITX) Stop(ctx context.Context) error {
	if !r.isExecuting.Load() {
		return ErrNotExecuting
	}

	r.processMu.RLock()
	process := r.process
	r.processMu.RUnlock()

	if process != nil {
		if err := process.Stop(ctx); err != nil {
			return ctxerrors.Wrap(err, "failed to stop process")
		}
	}

	return nil
}

// waitWithTimeout waits for process completion with manual timeout handling.
func (r *RPITX) waitWithTimeout(
	ctx context.Context,
	timeout time.Duration,
) error {
	errCh := make(chan error, 1)

	// Start waiting for process in goroutine
	go func() {
		errCh <- r.process.Wait()
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		// Process completed normally
		if err != nil {
			return ctxerrors.Wrap(err, "failed to wait for process")
		}

		return nil

	case <-time.After(timeout):
		// Timeout occurred - use graceful stop with timeout
		logrus.Debug("timeout reached, performing graceful stop")

		stopCtx, cancel := context.WithTimeout(
			ctx,
			gracefulStopTimeout,
		)

		defer cancel()

		err := r.Stop(stopCtx)
		if err != nil {
			logrus.WithError(err).
				Warn("failed to gracefully stop process after timeout")
		}

		// Wait for the stop to complete
		if err = <-errCh; err != nil {
			// Check if this was our expected timeout termination
			if errors.Is(err, commonerrors.ErrTerminated) ||
				errors.Is(err, commonerrors.ErrKilled) {
				return commonerrors.ErrTimeout
			}

			return ctxerrors.Wrap(err, "process failed after timeout stop")
		}

		return commonerrors.ErrTimeout
	}
}

// getMockExecCmd returns mock command and args for dev environment execution.
func (r *RPITX) getMockExecCmd(
	name ModuleName,
	args []string,
) (string, []string) {
	logrus.Debugf("preparing mock execution of module %s with args %s", name, args)

	// Build the mock command that echoes every second
	mockCmd := fmt.Sprintf(`
		while true; do
			echo "mocking execution of %s %s..."
			sleep 1
		done
	`, name, strings.Join(args, " "))

	// Return shell command and args
	return "sh", []string{"-c", mockCmd}
}
