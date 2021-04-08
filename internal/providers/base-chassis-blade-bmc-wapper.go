package providers

import (
	"fmt"
	"sync"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
)

type (
	baseChassisBladeBmcWrapper struct {
		username string
		password string
		host     string
		initOnce sync.Once
		bmc      devices.Cmc
	}
)

func (w *baseChassisBladeBmcWrapper) Close() error {
	if w.bmc != nil {
		return w.bmc.Close()
	}
	return nil
}

func (w *baseChassisBladeBmcWrapper) initBmcProvider() error {
	var err error

	w.initOnce.Do(func() {
		w.bmc, err = w.createBmcProvider()
	})

	return err
}

func (w *baseChassisBladeBmcWrapper) createBmcProvider() (devices.Cmc, error) {
	conn, err := discover.ScanAndConnect(w.host, w.username, w.password)
	if err != nil {
		return nil, fmt.Errorf("[baseChassisBladeBmcWrapper] Failed to setup BMC connection: %w", err)
	}

	bmc, ok := conn.(devices.Cmc)
	if !ok {
		return nil, fmt.Errorf("[baseChassisBladeBmcWrapper] Failed to cast the BMC connection to devices.Cmc")
	}

	if !bmc.IsActive() {
		return nil, fmt.Errorf("[baseChassisBladeBmcWrapper] Not an active device, actions can't be executed")
	}

	return bmc, nil
}
