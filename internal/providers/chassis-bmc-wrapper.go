package providers

type (
	ChassisBmcWrapper struct {
		*baseChassisBladeBmcWrapper
	}
)

func NewChassisBmcWrapper(username, password, host string) *ChassisBmcWrapper {
	return &ChassisBmcWrapper{
		baseChassisBladeBmcWrapper: &baseChassisBladeBmcWrapper{
			username: username,
			password: password,
			host:     host,
		},
	}
}

func (w *ChassisBmcWrapper) IsOn() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.IsOn()
}
func (w *ChassisBmcWrapper) PowerOn() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerOn()
}
func (w *ChassisBmcWrapper) PowerCycle() (bool, error) {
	if err := w.initBmcProvider(); err != nil {
		return false, err
	}
	return w.bmc.PowerCycle()
}
