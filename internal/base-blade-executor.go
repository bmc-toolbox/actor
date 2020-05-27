package internal

import (
	"fmt"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/providers"
)

type (
	BladeExecutorFactoryConfig struct {
		Username string
		Password string
	}

	baseBladeExecutor struct {
		bmc bladeBmcProvider
	}

	bladeBmcProvider interface {
		Close() error

		IsOnBlade(int) (bool, error)
		PowerOnBlade(int) (bool, error)
		PowerOffBlade(int) (bool, error)
		PowerCycleBlade(int) (bool, error)
		PowerCycleBmcBlade(int) (bool, error)
		ReseatBlade(int) (bool, error)

		PxeOnceBlade(int) (bool, error)

		FindBladePosition(string) (int, error)
	}
)

func newBaseBladeExecutor(username, password, host string) *baseBladeExecutor {
	return &baseBladeExecutor{bmc: providers.NewBladeBmcWrapper(username, password, host)}
}

func (e *baseBladeExecutor) Validate(action string) error {
	_, err := e.matchActionToFn(action)
	return err
}

func (e *baseBladeExecutor) matchActionToFn(action string) (func(int) (bool, error), error) {
	switch action {
	case actions.IsOn:
		return e.bmc.IsOnBlade, nil
	case actions.PowerOn:
		return e.bmc.PowerOnBlade, nil
	case actions.PowerOff:
		return e.bmc.PowerOffBlade, nil
	case actions.PowerCycle:
		return e.bmc.PowerCycleBlade, nil
	case actions.PowerCycleBmc:
		return e.bmc.PowerCycleBmcBlade, nil
	case actions.PxeOnce:
		return e.bmc.PxeOnceBlade, nil
	case actions.Reseat:
		return e.bmc.ReseatBlade, nil
	}

	return nil, fmt.Errorf("unknown action %q", action)
}

func (e *baseBladeExecutor) doAction(action string, bladePos int) actions.ActionResult {
	fn, err := e.matchActionToFn(action)
	if err != nil {
		return actions.NewActionResult(action, false, "failed", err)
	}

	status, err := fn(bladePos)
	if err != nil {
		return actions.NewActionResult(action, status, "failed", err)
	}
	return actions.NewActionResult(action, status, "ok", nil)
}

func (e *baseBladeExecutor) Cleanup() {
	if e.bmc != nil {
		e.bmc.Close()
		e.bmc = nil
	}
}
