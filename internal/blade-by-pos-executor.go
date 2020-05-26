package internal

import (
	"fmt"
	"strconv"

	"github.com/bmc-toolbox/actor/internal/executor"
)

type (
	BladeByPosExecutorFactory struct {
		config *BladeExecutorFactoryConfig
	}

	BladeByPosExecutor struct {
		*baseBladeExecutor
		bladePos int
	}
)

func NewBladeByPosExecutorFactory(config *BladeExecutorFactoryConfig) *BladeByPosExecutorFactory {
	return &BladeByPosExecutorFactory{config: config}
}

func (f *BladeByPosExecutorFactory) New(params map[string]interface{}) (executor.Executor, error) {
	if err := validateParam(params, paramHost, paramBladePosition); err != nil {
		return nil, fmt.Errorf("failed to validate params: %w", err)
	}

	bladePosStr := fmt.Sprintf("%v", params[paramBladePosition])
	bladePos, err := strconv.Atoi(bladePosStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameter %s from %q: %w", paramBladePosition, bladePosStr, err)
	}

	baseExecutor := newBaseBladeExecutor(f.config.Username, f.config.Password, fmt.Sprintf("%v", params[paramHost]))

	return &BladeByPosExecutor{baseBladeExecutor: baseExecutor, bladePos: bladePos}, nil
}

func (e *BladeByPosExecutor) Run(action string) executor.ActionResult {
	return e.doAction(action, e.bladePos)
}
