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
func Sleep(sleep string) (err error) {
	sleep = strings.Replace(sleep, "sleep ", "", 1)
	s, err := time.ParseDuration(sleep)
	if err != nil {
		return fmt.Errorf("error sleeping: %v", err)
	}
	time.Sleep(s)

	return err
}
