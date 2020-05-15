package internal

import (
	"errors"
	"fmt"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/executor"
	"github.com/bmc-toolbox/actor/internal/providers/ipmi"
	"github.com/bmc-toolbox/actor/internal/screenshot"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerrors "github.com/bmc-toolbox/bmclib/errors"
)

type (
	HostExecutorFactory struct {
		config *HostExecutorFactoryConfig
	}

	HostExecutor struct {
		config       *hostExecutorConfig
		bmc          hostBmcProvider
		screenshoter screenshot.BmcScreenshoter
		actionToFn   map[string]executor.ActionFn
	}

	HostExecutorFactoryConfig struct {
		IsS3     bool
		Username string
		Password string
	}

	hostExecutorConfig struct {
		*HostExecutorFactoryConfig
		host string
	}

	// this is abstraction over devices.Bmc and ipmi.Ipmi
	hostBmcProvider interface {
		Close() error

		IsOn() (bool, error)
		PowerOn() (bool, error)
		PowerOff() (bool, error)
		PowerCycle() (bool, error)
		PowerCycleBmc() (bool, error)

		PxeOnce() (bool, error)
	}
)

func NewHostExecutorFactory(config *HostExecutorFactoryConfig) *HostExecutorFactory {
	return &HostExecutorFactory{config: config}
}

func (f *HostExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if err := validateParam(params, paramHost); err != nil {
		return nil, fmt.Errorf("failed to create a new executor: %w", err)
	}

	config := &hostExecutorConfig{
		HostExecutorFactoryConfig: f.config,
		host:                      fmt.Sprintf("%v", params[paramHost]),
	}

	hostExecutor := &HostExecutor{config: config}
	hostExecutor.initFunctionMap()

	return hostExecutor, nil
}

func (e *HostExecutor) initFunctionMap() {
	e.actionToFn = map[string]executor.ActionFn{
		actions.IsOn:          e.doIsOn,
		actions.PowerOn:       e.doPowerOn,
		actions.PowerOff:      e.doPowerOff,
		actions.PowerCycle:    e.doPowerCycle,
		actions.PowerCycleBmc: e.doPowerCycleBmc,
		actions.PxeOnce:       e.doPxeOnce,
		actions.Screenshot:    e.doScreenshot,
	}
}

func (e *HostExecutor) ActionToFn(action string) (executor.ActionFn, error) {
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

func (e *HostExecutor) doSleep(action string) executor.ActionResult {
	if err := actions.Sleep(action); err != nil {
		return executor.NewActionResult(action, false, "failed", err)
	}
	return executor.NewActionResult(action, true, "ok", nil)
}

func (e *HostExecutor) doScreenshot() executor.ActionResult {
	if err := e.setupBmc(); err != nil {
		return executor.NewActionResult(actions.Screenshot, false, "failed", err)
	}

	if e.screenshoter == nil {
		err := fmt.Errorf("BMC provider not found")
		return executor.NewActionResult(actions.Screenshot, false, "failed", err)
	}

	if e.config.IsS3 {
		message, status, err := screenshot.S3(e.screenshoter, e.config.host)
		return executor.NewActionResult(actions.Screenshot, status, message, err)
	}
	message, status, err := screenshot.Local(e.screenshoter, e.config.host)
	return executor.NewActionResult(actions.Screenshot, status, message, err)
}

func (e *HostExecutor) doIsOn() executor.ActionResult {
	if err := e.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.IsOn, false, "failed", err)
	}

	return e.doAction(actions.IsOn, e.bmc.IsOn)
}

func (e *HostExecutor) doPowerOn() executor.ActionResult {
	if err := e.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOn, false, "failed", err)
	}

	return e.doAction(actions.PowerOn, e.bmc.PowerOn)
}

func (e *HostExecutor) doPowerOff() executor.ActionResult {
	if err := e.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOff, false, "failed", err)
	}

	return e.doAction(actions.PowerOff, e.bmc.PowerOff)
}

func (e *HostExecutor) doPowerCycle() executor.ActionResult {
	if err := e.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycle, false, "failed", err)
	}

	return e.doAction(actions.PowerCycle, e.bmc.PowerCycle)
}

func (e *HostExecutor) doPowerCycleBmc() executor.ActionResult {
	if err := e.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycleBmc, false, "failed", err)
	}

	return e.doAction(actions.PowerCycleBmc, e.bmc.PowerCycleBmc)
}

func (e *HostExecutor) doPxeOnce() executor.ActionResult {
	if err := e.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PxeOnce, false, "failed", err)
	}

	return e.doAction(actions.PxeOnce, e.bmc.PxeOnce)
}

func (e *HostExecutor) doAction(action string, fn func() (bool, error)) executor.ActionResult {
	status, err := fn()
	if err != nil {
		return executor.NewActionResult(action, status, "failed", err)
	}

	return executor.NewActionResult(action, status, "ok", nil)
}

// setupHostBmcProvider tries to setup a BMC provider with fallback to an IPMI
func (e *HostExecutor) setupHostBmcProvider() error {
	if err := e.setupBmc(); err != nil {
		// fall back to IPMI
		var errUH *bmcerrors.ErrUnsupportedHardware
		if errors.As(err, &errUH) || errors.Is(err, bmcerrors.ErrVendorNotSupported) {
			return e.setupIpmi()
		}

		return err
	}

	return nil
}

func (e *HostExecutor) setupBmc() error {
	if e.bmc != nil && e.screenshoter != nil {
		return nil
	}

	conn, err := discover.ScanAndConnect(e.config.host, e.config.Username, e.config.Password)
	if err != nil {
		return fmt.Errorf("failed to setup BMC connection: %w", err)
	}

	if bmc, ok := conn.(devices.Bmc); ok {
		e.bmc = bmc
		e.screenshoter = bmc
		return nil
	}

	return fmt.Errorf("failed to cast the BMC connection to devices.Bmc")
}

func (e *HostExecutor) setupIpmi() error {
	if e.bmc != nil {
		return nil
	}

	bmc, err := ipmi.New(e.config.host, e.config.Username, e.config.Password)
	if err != nil {
		return fmt.Errorf("failed to setup IPMI connection: %w", err)
	}

	e.bmc = bmc

	return nil
}

func (e *HostExecutor) Cleanup() {
	if e.bmc != nil {
		e.bmc.Close()
		e.bmc = nil
		e.screenshoter = nil
	}
}
