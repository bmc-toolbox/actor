package actions

import (
	"fmt"
	"net"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/ssh"
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

// SshCall execute the given command and returns a string with the output
func SshCall(client *ssh.Client, command string) (result string, err error) {
	session, err := client.NewSession()
	if err != nil {
		return result, err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), err
	}

	return string(output), err
}

// IsntLetterOrNumber check if the give rune is not a letter nor a number
func IsntLetterOrNumber(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

// SshBuildConfig builds a generic ssh config to use across all providers
func SshBuildConfig(username string, password string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 15 * time.Second,
	}
}

// SshBuildConfigCluster builds an ssh config to use with HP chassis
func SshBuildConfigCluster(username string, password string, master *bool) *ssh.ClientConfig {
	s := SshBuildConfig(username, password)
	s.BannerCallback = isHPChassisMaster(master)
	return s
}

// isHPChassisMaster inspect the HP chassis SSH banner to see if it's the active server
func isHPChassisMaster(master *bool) ssh.BannerCallback {
	return func(banner string) (err error) {
		*master = false
		if strings.Contains(banner, "Active") {
			*master = true
		}
		return err
	}
}
