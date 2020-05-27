package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/bmc-toolbox/actor/internal/actions"
)

type (
	SleepExecutorFactory struct {
	}

	SleepExecutor struct {
	}
)

func (f *SleepExecutorFactory) New(_ map[string]interface{}) (actions.Executor, error) {
	return &SleepExecutor{}, nil
}

func (e *SleepExecutor) Validate(action string) error {
	ok, err := isSleepAction(action)
	if !ok {
		return fmt.Errorf("%q is not a sleep action", action)
	}
	return err
}

func (e *SleepExecutor) Run(action string) actions.ActionResult {
	duration, err := parserDuration(action)
	if err != nil {
		return actions.NewActionResult(action, false, "failed", err)
	}

	time.Sleep(duration)

	return actions.NewActionResult(action, true, "ok", nil)
}

func (e *SleepExecutor) Cleanup() {
}

func parserDuration(sleepAction string) (time.Duration, error) {
	durationStr := strings.Replace(sleepAction, "sleep ", "", 1)
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parser duration in sleep action: %w", err)
	}

	return duration, nil
}

func isSleepAction(action string) (bool, error) {
	if strings.HasPrefix(action, "sleep") && strings.Contains(action, "sleep ") {
		if _, err := parserDuration(action); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}
