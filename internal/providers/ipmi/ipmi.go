package ipmi

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TODO: move it into bmclib
type Ipmi struct {
	Username    string
	Password    string
	Host        string
	ipmitoolBin string
}

func New(host string, username string, password string) (*Ipmi, error) {
	ipmi := &Ipmi{
		Username: username,
		Password: password,
		Host:     host,
	}

	ipmitoolBin, err := ipmi.findBin("ipmitoolBin")
	if err != nil {
		return nil, err
	}

	ipmi.ipmitoolBin = ipmitoolBin

	return ipmi, nil
}

func (i *Ipmi) run(command []string) (string, error) {
	args := []string{"-I", "lanplus", "-U", i.Username, "-E", "-H", i.Host}
	args = append(args, command...)

	cmd := exec.Command(i.ipmitoolBin, args...)
	cmd.Env = []string{fmt.Sprintf("IPMITOOL_PASSWORD=%s", i.Password)}

	out, err := cmd.CombinedOutput()

	return string(out), err
}

func (i *Ipmi) findBin(binary string) (string, error) {
	locations := []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/usr/local/sbin"}

	for _, path := range locations {
		lookup := path + "/" + binary
		fileInfo, err := os.Stat(path + "/" + binary)

		if err != nil {
			continue
		}

		if !fileInfo.IsDir() {
			return lookup, nil
		}
	}

	return "", fmt.Errorf("unable to find binary: %s", binary)
}

// PowerCycle reboots the machine via bmc
func (i *Ipmi) PowerCycle() (bool, error) {
	output, err := i.run([]string{"chassis", "power", "reset"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Chassis Power Control: Reset") {
		return true, err
	}

	return false, fmt.Errorf(output)
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *Ipmi) PowerCycleBmc() (bool, error) {
	output, err := i.run([]string{"mc", "reset", "cold"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Sent cold reset command to MC") {
		return true, err
	}

	return false, fmt.Errorf(output)
}

// PowerOn power on the machine via bmc
func (i *Ipmi) PowerOn() (bool, error) {
	status, err := i.IsOn()
	if err != nil {
		return false, err
	}

	if status {
		return false, fmt.Errorf("server is already on")
	}

	return i.powerOn()
}

// PowerOnForce power on the machine via bmc even when the machine is already on (Thanks HP!)
func (i *Ipmi) PowerOnForce() (bool, error) {
	return i.powerOn()
}

func (i *Ipmi) powerOn() (bool, error) {
	output, err := i.run([]string{"chassis", "power", "on"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Chassis Power Control: Up/On") {
		return true, err
	}

	return false, fmt.Errorf(output)
}

// PowerOff power off the machine via bmc
func (i *Ipmi) PowerOff() (bool, error) {
	status, err := i.IsOn()
	if err != nil {
		return false, err
	}

	if !status {
		return false, fmt.Errorf("server is already off")
	}

	output, err := i.run([]string{"chassis", "power", "off"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Chassis Power Control: Down/Off") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PxeOnceEfi makes the machine to boot via pxe once using EFI
func (i *Ipmi) PxeOnceEfi() (bool, error) {
	output, err := i.run([]string{"chassis", "bootdev", "pxe", "options=efiboot"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return i.PowerCycle()
	}

	return false, fmt.Errorf(output)
}

// PxeOnceMbr makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnceMbr() (bool, error) {
	output, err := i.run([]string{"chassis", "bootdev", "pxe"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return i.PowerCycle()
	}

	return false, fmt.Errorf(output)
}

// PxeOnce makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnce() (bool, error) {
	return i.PxeOnceEfi()
}

// IsOn tells if a machine is currently powered on
func (i *Ipmi) IsOn() (bool, error) {
	output, err := i.run([]string{"chassis", "power", "status"})
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Chassis Power is on") {
		return true, nil
	}

	if strings.Contains(output, "Chassis Power is off") {
		return false, nil
	}

	return false, fmt.Errorf(output)
}

// Close is a dummy connection to supply the interface
func (i *Ipmi) Close() error {
	return nil
}
