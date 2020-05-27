package internal

import (
	"fmt"

	"github.com/bmc-toolbox/actor/internal/executor"
)

type (
	BladeBySerialExecutorFactory struct {
		username string
		password string
	}

	BladeBySerialExecutor struct {
		*baseBladeExecutor
		bladeSerial string
	}
)

func NewBladeBySerialExecutorFactory(username, password string) *BladeBySerialExecutorFactory {
	return &BladeBySerialExecutorFactory{username: username, password: password}
}

func (f *BladeBySerialExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if err := validateParam(params, paramHost, paramBladeSerial); err != nil {
		return nil, fmt.Errorf("failed to validate params: %w", err)
	}

	baseExecutor := newBaseBladeExecutor(f.username, f.password, fmt.Sprintf("%v", params[paramHost]))
	bladeSerial := fmt.Sprintf("%v", params[paramBladeSerial])

	return &BladeBySerialExecutor{baseBladeExecutor: baseExecutor, bladeSerial: bladeSerial}, nil
}

func (e *BladeBySerialExecutor) Validate(action string) error {
	_, err := e.matchActionToFn(action)
	return err
}

func (e *BladeBySerialExecutor) Run(action string) executor.ActionResult {
	bladePos, err := e.bmc.FindBladePosition(e.bladeSerial)
	if err != nil {
		return executor.NewActionResult(action, false, "failed", err)
	}

	return e.doAction(action, bladePos)
}
