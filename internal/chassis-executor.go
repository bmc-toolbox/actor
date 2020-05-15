package internal

import (
	"fmt"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/executor"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
)

type (
	ChassisExecutorFactory struct {
		config *ChassisExecutorFactoryConfig
	}

	ChassisExecutor struct {
		config     *chassisExecutorConfig
		bmc        chassisBmcProvider
		actionToFn map[string]executor.ActionFn
	}

	ChassisExecutorFactoryConfig struct {
		Username string
		Password string
	}

	chassisExecutorConfig struct {
		*ChassisExecutorFactoryConfig
		host string
	}

	chassisBmcProvider interface {
		IsOn() (bool, error)
		PowerOn() (bool, error)
		PowerCycle() (bool, error)
		Close() error
	}
)

func NewChassisExecutorFactory(config *ChassisExecutorFactoryConfig) *ChassisExecutorFactory {
	return &ChassisExecutorFactory{config: config}
}

func (f *ChassisExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if err := validateParam(params, paramHost); err != nil {
		return nil, fmt.Errorf("failed to create a new executor: %w", err)
	}

	config := &chassisExecutorConfig{
		ChassisExecutorFactoryConfig: f.config,
		host:                         fmt.Sprintf("%v", params[paramHost]),
	}

	chassisExecutor := &ChassisExecutor{config: config}
	chassisExecutor.initFunctionMap()

	return chassisExecutor, nil
}

func (e *ChassisExecutor) initFunctionMap() {
	e.actionToFn = map[string]executor.ActionFn{
		actions.IsOn:       e.doIsOn,
		actions.PowerOn:    e.doPowerOn,
		actions.PowerCycle: e.doPowerCycle,
	}
}

func (e *ChassisExecutor) ActionToFn(action string) (executor.ActionFn, error) {
	if fn, ok := e.actionToFn[action]; ok {
		return fn, nil
	}

	ok, err := actions.IsSleepAction(action)
	if err != nil {
		return nil, fmt.Errorf("invalid action %q: %w", action, err)
	}

	if ok {
		return func() executor.ActionResult { return e.doSleep(action) }, nil
	}

	return nil, fmt.Errorf("unknown action %q", action)
}

func (e *ChassisExecutor) doSleep(action string) executor.ActionResult {
	if err := actions.Sleep(action); err != nil {
		return executor.NewActionResult(action, false, "failed", err)
	}
	return executor.NewActionResult(action, true, "ok", nil)
}

func (e *ChassisExecutor) doIsOn() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.IsOn, false, "failed", err)
	}

	return e.doAction(actions.IsOn, e.bmc.IsOn)
}

func (e *ChassisExecutor) doPowerOn() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOn, false, "failed", err)
	}

	return e.doAction(actions.PowerOn, e.bmc.PowerOn)
}

func (e *ChassisExecutor) doPowerCycle() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycle, false, "failed", err)
	}

	return e.doAction(actions.PowerCycle, e.bmc.PowerCycle)
}

func (e *ChassisExecutor) doAction(action string, fn func() (bool, error)) executor.ActionResult {
	status, err := fn()
	if err != nil {
		return executor.NewActionResult(action, status, "failed", err)
	}

	return executor.NewActionResult(action, status, "ok", nil)
}

func (e *ChassisExecutor) setupBmcProvider() error {
	if e.bmc != nil {
		return nil
	}

	conn, err := discover.ScanAndConnect(e.config.host, e.config.Username, e.config.Password)
	if err != nil {
		return fmt.Errorf("failed to setup BMC connection: %w", err)
	}

	bmc, ok := conn.(devices.Cmc)
	if !ok {
		return fmt.Errorf("failed to cast the BMC connection to devices.Cmc")
	}

	if !bmc.IsActive() {
		return fmt.Errorf("it is not active device, actions cannot be executed")
	}

	e.bmc = bmc

	return nil
}

func (e *ChassisExecutor) Cleanup() {
	if e.bmc != nil {
		e.bmc.Close()
		e.bmc = nil
	}
}
