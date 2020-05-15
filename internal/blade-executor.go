package internal

import (
	"fmt"
	"strconv"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/executor"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
)

type (
	BladeExecutorFactory struct {
		config *BladeExecutorFactoryConfig
	}

	BladeExecutor struct {
		config     *bladeExecutorConfig
		bmc        bladeBmcProvider
		actionToFn map[string]executor.ActionFn
	}

	BladeExecutorFactoryConfig struct {
		Username string
		Password string
	}

	bladeExecutorConfig struct {
		*BladeExecutorFactoryConfig
		host     string
		bladePos int
	}

	bladeBmcProvider interface {
		PowerCycleBlade(int) (bool, error)
		IsOnBlade(int) (bool, error)
		PxeOnceBlade(int) (bool, error)
		PowerCycleBmcBlade(int) (bool, error)
		PowerOnBlade(int) (bool, error)
		PowerOffBlade(int) (bool, error)
		ReseatBlade(int) (bool, error)
		Close() error
	}
)

func NewBladeExecutorFactory(config *BladeExecutorFactoryConfig) *BladeExecutorFactory {
	return &BladeExecutorFactory{config: config}
}

func (f *BladeExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if err := validateParam(params, paramHost, paramBladePosition); err != nil {
		return nil, fmt.Errorf("failed to validate params: %w", err)
	}

	bladePosStr := fmt.Sprintf("%v", params[paramBladePosition])
	bladePos, err := strconv.Atoi(bladePosStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameter %s from %q: %w", paramBladePosition, bladePosStr, err)
	}

	config := &bladeExecutorConfig{
		BladeExecutorFactoryConfig: f.config,
		host:                       fmt.Sprintf("%v", params[paramHost]),
		bladePos:                   bladePos,
	}

	bladeExecutor := &BladeExecutor{config: config}
	bladeExecutor.initFunctionMap()

	return bladeExecutor, nil
}

func (e *BladeExecutor) initFunctionMap() {
	e.actionToFn = map[string]executor.ActionFn{
		actions.IsOn:          e.doIsOnBlade,
		actions.PowerOn:       e.doPowerOnBlade,
		actions.PowerOff:      e.doPowerOffBlade,
		actions.PowerCycle:    e.doPowerCycleBlade,
		actions.PowerCycleBmc: e.doPowerCycleBmcBlade,
		actions.PxeOnce:       e.doPxeOnceBlade,
		actions.Reseat:        e.doReseatBlade,
	}
}

func (e *BladeExecutor) ActionToFn(action string) (executor.ActionFn, error) {
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

func (e *BladeExecutor) doSleep(action string) executor.ActionResult {
	if err := actions.Sleep(action); err != nil {
		return executor.NewActionResult(action, false, "failed", err)
	}
	return executor.NewActionResult(action, true, "ok", nil)
}

func (e *BladeExecutor) doIsOnBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.IsOn, false, "failed", err)
	}

	return e.doAction(actions.IsOn, e.bmc.IsOnBlade)
}

func (e *BladeExecutor) doPowerOnBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOn, false, "failed", err)
	}

	return e.doAction(actions.PowerOn, e.bmc.PowerOnBlade)
}

func (e *BladeExecutor) doPowerOffBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOff, false, "failed", err)
	}

	return e.doAction(actions.PowerOff, e.bmc.PowerOffBlade)
}

func (e *BladeExecutor) doPowerCycleBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycle, false, "failed", err)
	}

	return e.doAction(actions.PowerCycle, e.bmc.PowerCycleBlade)
}

func (e *BladeExecutor) doPowerCycleBmcBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycleBmc, false, "failed", err)
	}

	return e.doAction(actions.PowerCycleBmc, e.bmc.PowerCycleBmcBlade)
}

func (e *BladeExecutor) doPxeOnceBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PxeOnce, false, "failed", err)
	}

	return e.doAction(actions.PxeOnce, e.bmc.PxeOnceBlade)
}

func (e *BladeExecutor) doReseatBlade() executor.ActionResult {
	if err := e.setupBmcProvider(); err != nil {
		return executor.NewActionResult(actions.Reseat, false, "failed", err)
	}

	return e.doAction(actions.Reseat, e.bmc.ReseatBlade)
}

func (e *BladeExecutor) doAction(action string, fn func(int) (bool, error)) executor.ActionResult {
	status, err := fn(e.config.bladePos)
	if err != nil {
		return executor.NewActionResult(action, status, "failed", err)
	}

	return executor.NewActionResult(action, status, "ok", nil)
}

func (e *BladeExecutor) setupBmcProvider() error {
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

func (e *BladeExecutor) Cleanup() {
	if e.bmc != nil {
		e.bmc.Close()
		e.bmc = nil
	}
}
