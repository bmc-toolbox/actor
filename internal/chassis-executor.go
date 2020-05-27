package internal

import (
	"fmt"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/executor"
	"github.com/bmc-toolbox/actor/internal/providers"
)

type (
	ChassisExecutorFactory struct {
		username string
		password string
	}

	ChassisExecutor struct {
		bmc chassisBmcProvider
	}

	chassisBmcProvider interface {
		IsOn() (bool, error)
		PowerOn() (bool, error)
		PowerCycle() (bool, error)
		Close() error
	}
)

func NewChassisExecutorFactory(username, password string) *ChassisExecutorFactory {
	return &ChassisExecutorFactory{username: username, password: password}
}

func (f *ChassisExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if err := validateParam(params, paramHost); err != nil {
		return nil, fmt.Errorf("failed to create a new executor: %w", err)
	}

	host := fmt.Sprintf("%v", params[paramHost])

	return &ChassisExecutor{bmc: providers.NewChassisBmcWrapper(f.username, f.password, host)}, nil
}

func (e *ChassisExecutor) Validate(action string) error {
	_, err := e.matchActionToFn(action)
	return err
}

func (e *ChassisExecutor) Run(action string) executor.ActionResult {
	return e.doBMCAction(action)
}

func (e *ChassisExecutor) matchActionToFn(action string) (func() (bool, error), error) {
	switch action {
	case actions.IsOn:
		return e.bmc.IsOn, nil
	case actions.PowerOn:
		return e.bmc.PowerOn, nil
	case actions.PowerCycle:
		return e.bmc.PowerCycle, nil
	}

	return nil, fmt.Errorf("unknown action %q", action)
}

func (e *ChassisExecutor) doBMCAction(action string) executor.ActionResult {
	fn, err := e.matchActionToFn(action)
	if err != nil {
		return executor.NewActionResult(action, false, "failed", err)
	}

	status, err := fn()
	if err != nil {
		return executor.NewActionResult(action, status, "failed", err)
	}
	return executor.NewActionResult(action, status, "ok", nil)
}

func (e *ChassisExecutor) Cleanup() {
	e.bmc.Close()
}
