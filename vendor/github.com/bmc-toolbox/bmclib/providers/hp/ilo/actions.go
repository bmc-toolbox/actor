package ilo

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// PowerCycle reboots the machine via bmc
func (i *Ilo) PowerCycle() (bool, error) {
	output, err := i.sshClient.Run("power reset")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Server power off") {
		return i.PowerOn()
	}

	if strings.Contains(output, "Server resetting") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *Ilo) PowerCycleBmc() (bool, error) {
	output, err := i.sshClient.Run("reset /map1")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Resetting iLO") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOn power on the machine via bmc
func (i *Ilo) PowerOn() (bool, error) {
	output, err := i.sshClient.Run("power on")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Server powering on") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOff power off the machine via bmc
func (i *Ilo) PowerOff() (bool, error) {
	output, err := i.sshClient.Run("power off hard")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Forcing server") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PxeOnce makes the machine to boot via pxe once
func (i *Ilo) PxeOnce() (bool, error) {
	p, err := ipmi.New(i.username, i.password, i.ip)
	if err != nil {
		return false, err
	}
	// PXE using UEFI does't work for some models directly.
	// It only works if you PXE, PowerCycle, and PowerOn.
	if _, err = p.PxeOnceEfi(context.Background()); err != nil {
		return false, err
	}

	if _, err := p.ForceRestart(context.Background()); err != nil {
		return false, err
	}

	return p.PowerOn(context.Background())
}

// IsOn tells if a machine is currently powered on
func (i *Ilo) IsOn() (bool, error) {
	output, err := i.sshClient.Run("power")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "currently: On") {
		return true, nil
	}

	if strings.Contains(output, "currently: Off") {
		return false, nil
	}

	return false, fmt.Errorf(output)
}

// UpdateFirmware updates the bmc firmware
func (i *Ilo) UpdateFirmware(source, file string) (bool, string, error) {
	cmd := fmt.Sprintf("load /map1/firmware1 -source %s/%s", source, file)
	output, err := i.sshClient.Run(cmd)
	if err != nil {
		return false, "", fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Resetting iLO") {
		return true, output, nil
	}

	return false, output, fmt.Errorf(output)
}

func (i *Ilo) CheckFirmwareVersion() (version string, err error) {
	output, err := i.sshClient.Run("show /map1/firmware1")
	if err != nil {
		return "", fmt.Errorf("output: %q: %w", output, err)
	}

	re := regexp.MustCompile(`.*version=((\d+\.)+\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("unexpected output: %q", output)
}
