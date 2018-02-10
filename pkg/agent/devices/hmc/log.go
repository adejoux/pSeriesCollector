package hmc

// Debugf info
func (d *HMCServer) Debugf(expr string, vars ...interface{}) {
	expr2 := "HMCServer [" + d.cfg.ID + "] " + expr
	d.log.Debugf(expr2, vars...)
}

// Infof info
func (d *HMCServer) Infof(expr string, vars ...interface{}) {
	expr2 := "HMCServer [" + d.cfg.ID + "] " + expr
	d.log.Infof(expr2, vars...)
}

// Errorf info
func (d *HMCServer) Errorf(expr string, vars ...interface{}) {
	expr2 := "HMCServer [" + d.cfg.ID + "] " + expr
	d.log.Errorf(expr2, vars...)
}

// Warnf log warn data
func (d *HMCServer) Warnf(expr string, vars ...interface{}) {
	expr2 := "HMCServer [" + d.cfg.ID + "] " + expr
	d.log.Warnf(expr2, vars...)
}
