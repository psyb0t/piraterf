package commander

import (
	"context"
	"errors"
	"os/exec"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

type executionContext struct {
	ctx  context.Context //nolint:containedctx
	cmd  *exec.Cmd
	name string
	args []string
}

func (c *commander) newExecutionContext(
	ctx context.Context,
	name string,
	args []string,
	opts *Options,
) *executionContext {
	cmd := c.createCmd(ctx, name, args, opts)

	return &executionContext{
		ctx:  ctx,
		cmd:  cmd,
		name: name,
		args: args,
	}
}

func (ec *executionContext) handleExecutionError(err error) error {
	if err == nil {
		logrus.Debugf("command completed successfully: %s %v", ec.name, ec.args)

		return nil
	}

	logrus.Debugf("command execution failed: %s %v - error: %v", ec.name, ec.args, err)

	if errors.Is(err, context.DeadlineExceeded) {
		return commonerrors.ErrTimeout
	}

	if errors.Is(ec.ctx.Err(), context.DeadlineExceeded) &&
		isKilledBySignal(err) {
		return commonerrors.ErrTimeout
	}

	return ctxerrors.Wrap(err, "command failed")
}
