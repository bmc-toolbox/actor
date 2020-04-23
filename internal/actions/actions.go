package actions

import (
	"fmt"
	"strings"
	"time"
)

const (
	PowerOff      = "poweroff"
	PowerOn       = "poweron"
	PowerCycle    = "powercycle"
	HardReset     = "hardreset"
	Reseat        = "reseat"
	IsOn          = "ison"
	PowerCycleBmc = "powercyclebmc"
	PxeOnce       = "pxeonce"
	PxeOnceMBR    = "pxeoncembr"
	PxeOnceEFI    = "pxeonceefi"
	Screenshot    = "screenshot"
)

// Sleep transforms a sleep statement in a sleep-able time
func Sleep(action string) error {
	duration, err := parserDuration(action)
	if err != nil {
		return err
	}

	time.Sleep(duration)

	return err
}

// IsSleepAction checks if an action is the sleep action
func IsSleepAction(action string) (bool, error) {
	if strings.HasPrefix(action, "sleep") && strings.Contains(action, "sleep ") {
		if _, err := parserDuration(action); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}

func parserDuration(sleepAction string) (time.Duration, error) {
	durationStr := strings.Replace(sleepAction, "sleep ", "", 1)
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parser duration in sleep action: %w", err)
	}

	return duration, nil
}
