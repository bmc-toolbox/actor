package internal

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/providers"
	"github.com/bmc-toolbox/actor/internal/screenshot"
)

type (
	HostExecutorFactory struct {
		isS3Enabled bool
		username    string
		password    string
	}

	hostExecutor struct {
		bmc         bmcProvider
		host        string
		isS3Enabled bool
	}

	bmcProvider interface {
		Close(context.Context) error

		IsOn() (bool, error)
		PowerOn() (bool, error)
		PowerOff() (bool, error)
		PowerCycle() (bool, error)
		PowerCycleBmc() (bool, error)

		PxeOnce() (bool, error)

		// TODO: it looks like the screenshot's stuff shouldn't be in `hostExecutor`
		Screenshot() ([]byte, string, error)
		HardwareType() string
	}
)

func NewHostExecutorFactory(username, password string, isS3Enabled bool) *HostExecutorFactory {
	return &HostExecutorFactory{username: username, password: password, isS3Enabled: isS3Enabled}
}

func (f *HostExecutorFactory) New(params map[string]interface{}) (actions.Executor, error) {
	if err := validateParam(params, paramHost); err != nil {
		return nil, fmt.Errorf("failed to create a new executor: %w", err)
	}

	host := fmt.Sprintf("%v", params[paramHost])

	hostExecutor := &hostExecutor{
		bmc:         providers.NewServerBmcWrapper(f.username, f.password, host),
		host:        host,
		isS3Enabled: f.isS3Enabled,
	}

	return hostExecutor, nil
}

func (e *hostExecutor) Validate(action string) error {
	_, err := e.matchServerActionToFn(action)
	if err == nil {
		return nil
	}

	_, err = e.matchScreenshotActionToFn(action)
	return err
}

func (e *hostExecutor) Run(action string) actions.ActionResult {
	return e.doAction(action)
}

func (e *hostExecutor) matchServerActionToFn(action string) (func() (bool, error), error) {
	switch action {
	case actions.IsOn:
		return e.bmc.IsOn, nil
	case actions.PowerOn:
		return e.bmc.PowerOn, nil
	case actions.PowerOff:
		return e.bmc.PowerOff, nil
	case actions.PowerCycle:
		return e.bmc.PowerCycle, nil
	case actions.PowerCycleBmc:
		return e.bmc.PowerCycleBmc, nil
	case actions.PxeOnce:
		return e.bmc.PxeOnce, nil
	}

	return nil, fmt.Errorf("unknown action %q", action)
}

func (e *hostExecutor) matchScreenshotActionToFn(action string) (func() (string, bool, error), error) {
	if action == actions.Screenshot {
		return e.doScreenshot, nil
	}

	return nil, fmt.Errorf("unknown action %q", action)
}

func (e *hostExecutor) doScreenshot() (string, bool, error) {
	if e.isS3Enabled {
		return screenshot.S3(e.bmc, e.host)
	}
	return screenshot.Local(e.bmc, e.host)
}

func (e *hostExecutor) doAction(action string) actions.ActionResult {
	serverFn, err := e.matchServerActionToFn(action)
	if err == nil {
		return e.doServerFn(action, serverFn)
	}

	screenshotFn, err := e.matchScreenshotActionToFn(action)
	if err == nil {
		return e.doScreenshotFn(action, screenshotFn)
	}

	return actions.NewActionResult(action, false, "failed", err)
}

func (e *hostExecutor) doScreenshotFn(action string, screenshotFn func() (string, bool, error)) actions.ActionResult {
	message, status, err := screenshotFn()
	if err != nil {
		return actions.NewActionResult(action, status, message, err)
	}
	return actions.NewActionResult(action, status, message, nil)
}

func (e *hostExecutor) doServerFn(action string, serverFn func() (bool, error)) actions.ActionResult {
	status, err := serverFn()
	if err != nil {
		return actions.NewActionResult(action, status, "failed", err)
	}
	return actions.NewActionResult(action, status, "ok", nil)
}

func (e *hostExecutor) Cleanup() {
	if e.bmc != nil {
		_ = e.bmc.Close(context.TODO())
	}
}
