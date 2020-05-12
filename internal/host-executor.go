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
	metrics "github.com/bmc-toolbox/gin-go-metrics"
)

const (
	paramHost = "host"
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

func (h *HostExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if _, ok := params[paramHost]; !ok {
		return nil, fmt.Errorf("failed to create a new executor: no required parameter %q", paramHost)
	}

	config := &hostExecutorConfig{
		HostExecutorFactoryConfig: h.config,
		host:                      fmt.Sprintf("%v", params[paramHost]),
	}

	hostExecutor := &HostExecutor{config: config}
	hostExecutor.initFunctionMap()

	return hostExecutor, nil
}

func (h *HostExecutor) initFunctionMap() {
	h.actionToFn = map[string]executor.ActionFn{
		actions.IsOn:          h.doIsOn,
		actions.PowerOn:       h.doPowerOn,
		actions.PowerOff:      h.doPowerOff,
		actions.PowerCycle:    h.doPowerCycle,
		actions.PowerCycleBmc: h.doPowerCycleBmc,
		actions.PxeOnce:       h.doPxeOnce,
		actions.Screenshot:    h.doScreenshot,
	}
}

func (h *HostExecutor) ActionToFn(action string) (executor.ActionFn, error) {
	if fn, ok := h.actionToFn[action]; ok {
		return fn, nil
	}

	ok, err := actions.IsSleepAction(action)
	if err != nil {
		return nil, fmt.Errorf("invalid action %q: %w", action, err)
	}

	if ok {
		return func() executor.ActionResult { return h.doSleep(action) }, nil
	}

	return nil, fmt.Errorf("unknown action %q", action)
}

func (h *HostExecutor) doSleep(action string) executor.ActionResult {
	if err := actions.Sleep(action); err != nil {
		return executor.NewActionResult(action, false, "failed", err)
	}
	return executor.NewActionResult(action, true, "ok", nil)
}

func (h *HostExecutor) doScreenshot() executor.ActionResult {
	if err := h.setupBmc(); err != nil {
		return executor.NewActionResult(actions.Screenshot, false, "failed", err)
	}

	if h.screenshoter == nil {
		err := fmt.Errorf("BMC provider not found")
		return executor.NewActionResult(actions.Screenshot, false, "failed", err)
	}

	if h.config.IsS3 {
		message, status, err := screenshot.S3(h.screenshoter, h.config.host)
		return executor.NewActionResult(actions.Screenshot, status, message, err)
	}
	message, status, err := screenshot.Local(h.screenshoter, h.config.host)
	return executor.NewActionResult(actions.Screenshot, status, message, err)
}

func (h *HostExecutor) doIsOn() executor.ActionResult {
	if err := h.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.IsOn, false, "failed", err)
	}

	return h.doAction(actions.IsOn, h.bmc.IsOn)
}

func (h *HostExecutor) doPowerOn() executor.ActionResult {
	if err := h.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOn, false, "failed", err)
	}

	return h.doAction(actions.PowerOn, h.bmc.PowerOn)
}

func (h *HostExecutor) doPowerOff() executor.ActionResult {
	if err := h.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerOff, false, "failed", err)
	}

	return h.doAction(actions.PowerOff, h.bmc.PowerOff)
}

func (h *HostExecutor) doPowerCycle() executor.ActionResult {
	if err := h.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycle, false, "failed", err)
	}

	return h.doAction(actions.PowerCycle, h.bmc.PowerCycle)
}

func (h *HostExecutor) doPowerCycleBmc() executor.ActionResult {
	if err := h.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PowerCycleBmc, false, "failed", err)
	}

	return h.doAction(actions.PowerCycleBmc, h.bmc.PowerCycleBmc)
}

func (h *HostExecutor) doPxeOnce() executor.ActionResult {
	if err := h.setupHostBmcProvider(); err != nil {
		return executor.NewActionResult(actions.PxeOnce, false, "failed", err)
	}

	return h.doAction(actions.PxeOnce, h.bmc.PxeOnce)
}

func (h *HostExecutor) doAction(action string, fn func() (bool, error)) executor.ActionResult {
	status, err := fn()
	if err != nil {
		return executor.NewActionResult(action, status, "failed", err)
	}

	return executor.NewActionResult(action, status, "ok", nil)
}

// setupHostBmcProvider tries to setup a BMC provider with fallback to an IPMI
func (h *HostExecutor) setupHostBmcProvider() error {
	if err := h.setupBmc(); err != nil {
		// fall back to IPMI
		var errUH *bmcerrors.ErrUnsupportedHardware
		if errors.As(err, &errUH) || errors.Is(err, bmcerrors.ErrVendorNotSupported) {
			return h.setupIpmi()
		}

		return err
	}

	return nil
}

func (h *HostExecutor) setupBmc() error {
	if h.bmc != nil && h.screenshoter != nil {
		return nil
	}

	conn, err := discover.ScanAndConnect(h.config.host, h.config.Username, h.config.Password)
	if err != nil {
		metrics.IncrCounter([]string{"errors", "bmc", "connect_fail"}, 1)
		return fmt.Errorf("failed to setup BMC connection: %w", err)
	}

	if bmc, ok := conn.(devices.Bmc); ok {
		h.bmc = bmc
		h.screenshoter = bmc
		return nil
	}

	return fmt.Errorf("failed to cast the BMC connection to devices.Bmc")
}

func (h *HostExecutor) setupIpmi() error {
	if h.bmc != nil {
		return nil
	}

	bmc, err := ipmi.New(h.config.host, h.config.Username, h.config.Password)
	if err != nil {
		metrics.IncrCounter([]string{"errors", "bmc", "ipmi_setup"}, 1)
		return fmt.Errorf("failed to setup IPMI connection: %w", err)
	}

	h.bmc = bmc

	return nil
}

func (h *HostExecutor) Cleanup() {
	if h.bmc != nil {
		h.bmc.Close()
		h.bmc = nil
		h.screenshoter = nil
	}
}
