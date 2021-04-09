package ipmi

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Ipmi struct {
	Username string
	Password string
	Host     string
	ipmitool string
}

func New(username string, password string, host string) (ipmi *Ipmi, err error) {
	ipmi = &Ipmi{
		Username: username,
		Password: password,
		Host:     host,
	}

	ipmi.ipmitool, err = exec.LookPath("ipmitool")
	if err != nil {
		return nil, err
	}
	return ipmi, nil
}

func (i *Ipmi) run(ctx context.Context, command []string) (output string, err error) {
	ipmiArgs := []string{"-I", "lanplus", "-U", i.Username, "-E", "-N", "5"}
	if strings.Contains(i.Host, ":") {
		host, port, err := net.SplitHostPort(i.Host)
		if err == nil {
			ipmiArgs = append(ipmiArgs, "-H", host, "-p", port)
		}
	} else {
		ipmiArgs = append(ipmiArgs, "-H", i.Host)
	}

	ipmiArgs = append(ipmiArgs, command...)
	cmd := exec.CommandContext(ctx, i.ipmitool, ipmiArgs...)
	cmd.Env = []string{fmt.Sprintf("IPMITOOL_PASSWORD=%s", i.Password)}
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), errors.Wrap(ctx.Err(), "[run, context.DeadlineExceeded]")
	}
	return string(out), errors.Wrap(err, "[run] Output: " + string(out) + "|" + i.ipmitool + " " + strings.Join(ipmiArgs, " "))
}

// Reboot the machine via BMC
func (i *Ipmi) PowerCycle(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "status"})
	if err != nil {
		return false, fmt.Errorf("[PowerCycle (status) Error] %v: %v", err, output)
	}

	command := "on"
	reply := "Up/On"
	if strings.HasPrefix(output, "Chassis Power is on") {
		command = "cycle"
		reply = "Cycle"
	} else if !strings.HasPrefix(output, "Chassis Power is off") {
		return false, fmt.Errorf("[PowerCycle (unexpected output)] %v", output)
	}

	output, err = i.run(ctx, []string{"chassis", "power", command})
	if err != nil {
		return false, fmt.Errorf("[PowerCycle (%v) Error] %v: %v", command, err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: " + reply) {
		return true, nil
	}
	return false, fmt.Errorf("[PowerCycle %v (unexpected output)] %v", command, output)
}

// Reset the machine via BMC
func (i *Ipmi) PowerReset(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "reset"})
	if err != nil {
		return false, fmt.Errorf("[PowerReset Error] %v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Reset") {
		return true, nil
	}
	return false, fmt.Errorf("[PowerReset (unexpected output)] %v", output)
}

// Reboot the BMC we are connected to
func (i *Ipmi) PowerCycleBmc(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"mc", "reset", "cold"})
	if err != nil {
		return false, fmt.Errorf("[PowerCycleBmc Error] %v: %v", err, output)
	}

	if strings.HasPrefix(output, "Sent cold reset command to MC") {
		return true, nil
	}
	return false, fmt.Errorf("[PowerCycleBmc (unexpected output)] %v", output)
}

// Reset the BMC we are connected to
func (i *Ipmi) PowerResetBmc(ctx context.Context, resetType string) (ok bool, err error) {
	output, err := i.run(ctx, []string{"mc", "reset", strings.ToLower(resetType)})
	if err != nil {
		return false, fmt.Errorf("[PowerResetBmc Error] %v: %v", err, output)
	}

	if strings.HasPrefix(output, fmt.Sprintf("Sent %v reset command to MC", strings.ToLower(resetType))) {
		return true, nil
	}
	return false, fmt.Errorf("[PowerResetBmc (unexpected output)] %v", output)
}

// Power the machine on via BMC
func (i *Ipmi) PowerOn(ctx context.Context) (status bool, err error) {
	s, err := i.IsOn(ctx)
	if err != nil {
		return false, fmt.Errorf("[PowerOn (IsOn) Error] %v", err)
	}

	if s {
		return false, fmt.Errorf("[PowerOn Warning] Server is already powered on!")
	}

	output, err := i.run(ctx, []string{"chassis", "power", "on"})
	if err != nil {
		return false, fmt.Errorf("[PowerOn Error] %v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Up/On") {
		return true, nil
	}
	return false, fmt.Errorf("[PowerOn (unexpected output)] %v", output)
}

// Power the machine off via BMC
func (i *Ipmi) PowerOff(ctx context.Context) (status bool, err error) {
	s, err := i.IsOn(ctx)
	if err != nil {
		return false, fmt.Errorf("[PowerOff (IsOn) Error] %v", err)
	}

	if !s {
		return false, fmt.Errorf("[PowerOff Warning] Server is already powered off!")
	}

	output, err := i.run(ctx, []string{"chassis", "power", "off"})
	if strings.Contains(output, "Chassis Power Control: Down/Off") {
		return true, nil
	}
	return false, fmt.Errorf("[PowerOff (unexpected output)] %v", output)
}

// Set the next boot device with options
func (i *Ipmi) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	var atLeastOneOptionSelected bool
	ipmiCmd := []string{"chassis", "bootdev", strings.ToLower(bootDevice)}
	var opts []string
	if setPersistent {
		opts = append(opts, "persistent")
		atLeastOneOptionSelected = true
	}
	if efiBoot {
		opts = append(opts, "efiboot")
		atLeastOneOptionSelected = true
	}
	if atLeastOneOptionSelected {
		optsJoined := strings.Join(opts, ",")
		optsFull := fmt.Sprintf("options=%v", optsJoined)
		ipmiCmd = append(ipmiCmd, optsFull)
	}

	output, err := i.run(ctx, ipmiCmd)
	if err != nil {
		return false, fmt.Errorf("[BootDeviceSet Error] %v: %v", err, output)
	}

	if strings.Contains(output, fmt.Sprintf("Set Boot Device to %v", strings.ToLower(bootDevice))) {
		return true, nil
	}
	return false, fmt.Errorf("[BootDeviceSet (unexpected output)] %v", output)
}

// Boot the machine via PXE once using EFI
func (i *Ipmi) PxeOnceEfi(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "bootdev", "pxe", "options=efiboot"})
	if err != nil {
		return false, fmt.Errorf("[PxeOnceEfi Error] %v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, nil
	}
	return false, fmt.Errorf("[PxeOnceEfi (unexpected output)] %v", output)
}
// Boot the machine via PXE once using MBR
func (i *Ipmi) PxeOnceMbr(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "bootdev", "pxe"})
	if err != nil {
		return false, fmt.Errorf("[PxeOnceMbr Error] %v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, nil
	}
	return false, fmt.Errorf("[PxeOnceMbr (unexpected output)] %v", output)
}
// The default is to PXE-boot via MBR
func (i *Ipmi) PxeOnce(ctx context.Context) (status bool, err error) {
	return i.PxeOnceMbr(ctx)
}

// Is the machine currently powered on?
func (i *Ipmi) IsOn(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "status"})
	if err != nil {
		return false, fmt.Errorf("[IsOn Error] %v: %v", err, output)
	}

	if strings.Contains(output, "Chassis Power is on") {
		return true, nil
	}
	return false, fmt.Errorf("[IsOn (unexpected output)] %v", output)
}

// Return the current power state of the machine
func (i *Ipmi) PowerState(ctx context.Context) (state string, err error) {
	return i.run(ctx, []string{"chassis", "power", "status"})
}

// List all BMC users
func (i *Ipmi) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	output, err := i.run(ctx, []string{"user", "list"})
	if err != nil {
		return users, errors.Wrap(err, "[ReadUsers] Error getting user list!")
	}

	header := map[int]string{}
	firstLine := true
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if firstLine {
			firstLine = false
			for x := 0; x < 5; x++ {
				header[x] = line[x]
			}
			continue
		}
		entry := map[string]string{}
		if line[1] != "true" {
			for x := 0; x < 5; x++ {
				entry[header[x]] = line[x]
			}
			users = append(users, entry)
		}
	}

	return users, err
}
