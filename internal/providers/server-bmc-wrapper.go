package providers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bmc-toolbox/actor/internal/providers/ipmi"
	"github.com/bmc-toolbox/actor/internal/screenshot"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerrors "github.com/bmc-toolbox/bmclib/errors"
)

type (
	ServerBmcWrapper struct {
		username     string
		password     string
		host         string
		initOnce     sync.Once
		bmc          serverBmcProvider
		screenshoter screenshot.BmcScreenshoter
	}

	// this is abstraction over devices.Bmc and ipmi.Ipmi
	serverBmcProvider interface {
		Close() error

		IsOn() (bool, error)
		PowerOn() (bool, error)
		PowerOff() (bool, error)
		PowerCycle() (bool, error)
		PowerCycleBmc() (bool, error)

		PxeOnce() (bool, error)
	}
)

func NewServerBmcWrapper(username, password, host string) *ServerBmcWrapper {
	return &ServerBmcWrapper{
		username: username,
		password: password,
		host:     host,
	}
}

func (w *ServerBmcWrapper) initBmcProvider() error {
	var err error

	w.initOnce.Do(func() {
		err = w.createBmcProviderWithFallback()
	})

	return err
}

func (w *ServerBmcWrapper) IsOn() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.IsOn()
}

func (w *ServerBmcWrapper) PowerOn() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerOn()
}

func (w *ServerBmcWrapper) PowerOff() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerOff()
}

func (w *ServerBmcWrapper) PowerCycle() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerCycle()
}

func (w *ServerBmcWrapper) PowerCycleBmc() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerCycleBmc()
}

func (w *ServerBmcWrapper) PxeOnce() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PxeOnce()
}

func (w *ServerBmcWrapper) Screenshot() ([]byte, string, error) {
	if err := w.initBmcProvider(); err != nil {
		return nil, "", err
	}

	if w.screenshoter == nil {
		return nil, "", fmt.Errorf("BMC provider does not support the method Screenshot()")
	}

	return w.screenshoter.Screenshot()
}

func (w *ServerBmcWrapper) HardwareType() string {
	if err := w.initBmcProvider(); err != nil {
		panic("failed to initialize BMC provider")
	}

	if w.screenshoter == nil {
		panic("BMC provider does not support the method HardwareType()")
	}

	return w.screenshoter.HardwareType()
}

func (w *ServerBmcWrapper) Close() error {
	if w.bmc != nil {
		w.screenshoter = nil
		return w.bmc.Close()
	}
	return nil
}

func (w *ServerBmcWrapper) createBmcProviderWithFallback() error {
	bmc, err := w.createBmcProvider()
	if err == nil {
		w.bmc = bmc
		w.screenshoter = bmc
		return nil
	}

	// fall back to IPMI
	var errUH *bmcerrors.ErrUnsupportedHardware
	if errors.As(err, &errUH) || errors.Is(err, bmcerrors.ErrVendorNotSupported) {
		ipmiProvider, ipmiErr := w.createIpmiProvider()
		if ipmiErr != nil {
			return ipmiErr
		}
		w.bmc = ipmiProvider
		return nil
	}

	return err
}

func (w *ServerBmcWrapper) createBmcProvider() (devices.Bmc, error) {
	conn, err := discover.ScanAndConnect(w.host, w.username, w.password)
	if err != nil {
		return nil, fmt.Errorf("failed to setup BMC connection: %w", err)
	}

	if bmc, ok := conn.(devices.Bmc); ok {
		return bmc, nil
	}

	return nil, fmt.Errorf("failed to cast the BMC connection to devices.Bmc")
}

func (w *ServerBmcWrapper) createIpmiProvider() (*ipmi.Ipmi, error) {
	bmc, err := ipmi.New(w.host, w.username, w.password)
	if err != nil {
		return nil, fmt.Errorf("failed to setup IPMI connection: %w", err)
	}

	return bmc, nil
}
