package routes

import (
	"errors"
	"fmt"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/providers/ipmi"
	"github.com/bmc-toolbox/actor/internal/screenshot"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerrors "github.com/bmc-toolbox/bmclib/errors"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/sirupsen/logrus"
)

type (
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

	hostActionExecutor struct {
		host         string
		username     string
		password     string
		isS3         bool
		bmc          hostBmcProvider
		screenshoter screenshot.BmcScreenshoter
		logger       *logrus.Entry
		actionToFn   map[string]func() (response, error)
		plan         []planEntry
	}

	planEntry struct {
		actionFn func() (response, error)
		action   string
	}
)

func newHostActionExecutor(host, username, password string, isS3 bool, logger *logrus.Entry) *hostActionExecutor {
	e := &hostActionExecutor{
		host:     host,
		username: username,
		password: password,
		isS3:     isS3,
		logger:   logger,
	}

	e.initActionToFunction()

	return e
}

func (e *hostActionExecutor) buildExecutionPlan(actionSeq []string) error {
	if len(actionSeq) == 0 {
		return fmt.Errorf("action sequence is empty")
	}

	plan := make([]planEntry, 0)

	for _, action := range actionSeq {
		a := action // do not pass `action` into the function of planEntity

		if fn, ok := e.actionToFn[a]; ok {
			plan = append(plan, planEntry{fn, a})
			continue
		}

		ok, err := actions.IsSleepAction(a)
		if err != nil {
			metrics.IncrCounter([]string{"errors", "bmc", "invalid_action"}, 1)
			return fmt.Errorf("invalid action %q: %w", a, err)
		}

		if !ok {
			metrics.IncrCounter([]string{"errors", "bmc", "unknown_action", a}, 1)
			return fmt.Errorf("unkown action %q", a)
		}

		plan = append(plan, planEntry{func() (response, error) { return e.doSleep(a) }, a})
	}

	e.plan = plan

	return nil
}

func (e *hostActionExecutor) run() ([]response, error) {
	if len(e.plan) == 0 {
		return nil, fmt.Errorf("execution plan is empty")
	}

	resps := make([]response, 0)

	for _, entry := range e.plan {
		resp, err := entry.actionFn()
		resps = append(resps, resp)
		if err != nil {
			e.logger.WithField("action", entry.action).WithError(err).Warn("error carrying out action")
			metrics.IncrCounter([]string{"action", "bmc", "fail", entry.action}, 1)
			return resps, err
		}

		metrics.IncrCounter([]string{"action", "bmc", "success", entry.action}, 1)
	}

	return resps, nil
}

func (e *hostActionExecutor) initActionToFunction() {
	e.actionToFn = map[string]func() (response, error){
		actions.IsOn:          e.doIsOn,
		actions.PowerOn:       e.doPowerOn,
		actions.PowerOff:      e.doPowerOff,
		actions.PowerCycle:    e.doPowerCycle,
		actions.PowerCycleBmc: e.doPowerCycleBmc,
		actions.PxeOnce:       e.doPxeOnce,
		actions.Screenshot:    e.doScreenshot,
	}
}

func (e *hostActionExecutor) doSleep(action string) (response, error) {
	if err := actions.Sleep(action); err != nil {
		return newResponse(action, false, "failed", err), err
	}
	return newResponse(action, true, "ok", nil), nil
}

func (e *hostActionExecutor) doScreenshot() (response, error) {
	if err := e.setupBmc(); err != nil {
		return newResponse(actions.Screenshot, false, "failed", err), err
	}

	if e.screenshoter == nil {
		err := fmt.Errorf("BMC provider not found")
		return newResponse(actions.Screenshot, false, "failed", err), err
	}

	if e.isS3 {
		message, status, err := screenshot.S3(e.screenshoter, e.host)
		return newResponse(actions.Screenshot, status, message, err), err
	}
	message, status, err := screenshot.Local(e.screenshoter, e.host)
	return newResponse(actions.Screenshot, status, message, err), err
}

func (e *hostActionExecutor) doIsOn() (response, error) {
	if err := e.setupHostBmcProvider(); err != nil {
		return newResponse(actions.IsOn, false, "failed", err), err
	}

	return e.doAction(actions.IsOn, e.bmc.IsOn)
}

func (e *hostActionExecutor) doPowerOn() (response, error) {
	if err := e.setupHostBmcProvider(); err != nil {
		return newResponse(actions.PowerOn, false, "failed", err), err
	}

	return e.doAction(actions.PowerOn, e.bmc.PowerOn)
}

func (e *hostActionExecutor) doPowerOff() (response, error) {
	if err := e.setupHostBmcProvider(); err != nil {
		return newResponse(actions.PowerOff, false, "failed", err), err
	}

	return e.doAction(actions.PowerOff, e.bmc.PowerOff)
}

func (e *hostActionExecutor) doPowerCycle() (response, error) {
	if err := e.setupHostBmcProvider(); err != nil {
		return newResponse(actions.PowerCycle, false, "failed", err), err
	}

	return e.doAction(actions.PowerCycle, e.bmc.PowerCycle)
}

func (e *hostActionExecutor) doPowerCycleBmc() (response, error) {
	if err := e.setupHostBmcProvider(); err != nil {
		return newResponse(actions.PowerCycleBmc, false, "failed", err), err
	}

	return e.doAction(actions.PowerCycleBmc, e.bmc.PowerCycleBmc)
}

func (e *hostActionExecutor) doPxeOnce() (response, error) {
	if err := e.setupHostBmcProvider(); err != nil {
		return newResponse(actions.PxeOnce, false, "failed", err), err
	}

	return e.doAction(actions.PxeOnce, e.bmc.PxeOnce)
}

func (e *hostActionExecutor) doAction(action string, fn func() (bool, error)) (response, error) {
	status, err := fn()
	if err != nil {
		return newResponse(action, status, "failed", err), err
	}

	return newResponse(action, status, "ok", nil), nil
}

// setupHostBmcProvider tries to setup a BMC provider with fallback to an IPMI
func (e *hostActionExecutor) setupHostBmcProvider() error {
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

func (e *hostActionExecutor) setupBmc() error {
	if e.bmc != nil && e.screenshoter != nil {
		return nil
	}

	conn, err := discover.ScanAndConnect(e.host, e.username, e.password)
	if err != nil {
		metrics.IncrCounter([]string{"errors", "bmc", "connect_fail"}, 1)
		return fmt.Errorf("failed to setup BMC connection: %w", err)
	}

	if bmc, ok := conn.(devices.Bmc); ok {
		e.bmc = bmc
		e.screenshoter = bmc
		return nil
	}

	return fmt.Errorf("failed to cast the BMC connection to devices.Bmc")
}

func (e *hostActionExecutor) setupIpmi() error {
	if e.bmc != nil {
		return nil
	}

	bmc, err := ipmi.New(e.host, e.username, e.password)
	if err != nil {
		metrics.IncrCounter([]string{"errors", "bmc", "ipmi_setup"}, 1)
		return fmt.Errorf("failed to setup IPMI connection: %w", err)
	}

	e.bmc = bmc

	return nil
}

func (e *hostActionExecutor) cleanup() {
	if e.bmc != nil {
		e.bmc.Close()
		e.bmc = nil
		e.screenshoter = nil
	}
}
