package internal

import (
	"fmt"
	"strconv"

	"github.com/bmc-toolbox/actor/internal/actions"
)

type (
	BladeByPosExecutorFactory struct {
		username string
		password string
	}

	BladeByPosExecutor struct {
		*baseBladeExecutor
		bladePos int
	}
)

func NewBladeByPosExecutorFactory(username, password string) *BladeByPosExecutorFactory {
	return &BladeByPosExecutorFactory{username: username, password: password}
}

func (f *BladeByPosExecutorFactory) New(params map[string]interface{}) (actions.Executor, error) {
	if err := validateParam(params, paramHost, paramBladePosition); err != nil {
		return nil, fmt.Errorf("failed to validate params: %w", err)
	}

	bladePosStr := fmt.Sprintf("%v", params[paramBladePosition])
	bladePos, err := strconv.Atoi(bladePosStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameter %s from %q: %w", paramBladePosition, bladePosStr, err)
	}

	baseExecutor := newBaseBladeExecutor(f.username, f.password, fmt.Sprintf("%v", params[paramHost]))

	return &BladeByPosExecutor{baseBladeExecutor: baseExecutor, bladePos: bladePos}, nil
}

func (e *BladeByPosExecutor) Run(action string) actions.ActionResult {
	return e.doAction(action, e.bladePos)
}
