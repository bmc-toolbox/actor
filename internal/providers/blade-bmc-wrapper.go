package providers

import "sync"

type (
	BladeBmcWrapper struct {
		*baseChassisBladeBmcWrapper
		bladeSerialToPos map[string]int
		lock             sync.RWMutex
	}
)

func NewBladeBmcWrapper(username, password, host string) *BladeBmcWrapper {
	return &BladeBmcWrapper{
		baseChassisBladeBmcWrapper: &baseChassisBladeBmcWrapper{
			username: username,
			password: password,
			host:     host,
		},
		bladeSerialToPos: make(map[string]int),
		lock:             sync.RWMutex{},
	}
}

func (w *BladeBmcWrapper) IsOnBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.IsOnBlade(position)
}

func (w *BladeBmcWrapper) PowerOnBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerOnBlade(position)
}

func (w *BladeBmcWrapper) PowerOffBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerOffBlade(position)
}

func (w *BladeBmcWrapper) PowerCycleBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerCycleBlade(position)
}

func (w *BladeBmcWrapper) PowerCycleBmcBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerCycleBmcBlade(position)
}

func (w *BladeBmcWrapper) ReseatBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.ReseatBlade(position)
}

func (w *BladeBmcWrapper) PxeOnceBlade(position int) (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PxeOnceBlade(position)
}

func (w *BladeBmcWrapper) FindBladePosition(serial string) (int, error) {
	if err := w.initBmcProvider(); err != nil {
		return -1, err
	}

	w.lock.RLock()
	if position, ok := w.bladeSerialToPos[serial]; ok {
		w.lock.RUnlock()
		return position, nil
	}

	w.lock.RUnlock()

	w.lock.Lock()
	defer w.lock.Unlock()

	if position, ok := w.bladeSerialToPos[serial]; ok {
		return position, nil
	}

	position, err := w.bmc.FindBladePosition(serial)
	if err != nil {
		return -1, err
	}

	w.bladeSerialToPos[serial] = position

	return position, nil
}
