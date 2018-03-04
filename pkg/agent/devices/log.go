package devices

// Debugf info
func (b *Base) Debugf(expr string, vars ...interface{}) {
	expr2 := b.logPrefix + expr
	b.log.Debugf(expr2, vars...)
}

// Infof info
func (b *Base) Infof(expr string, vars ...interface{}) {
	expr2 := b.logPrefix + expr
	b.log.Infof(expr2, vars...)
}

// Errorf info
func (b *Base) Errorf(expr string, vars ...interface{}) {
	expr2 := b.logPrefix + expr
	b.log.Errorf(expr2, vars...)
}

// Warnf log warn data
func (b *Base) Warnf(expr string, vars ...interface{}) {
	expr2 := b.logPrefix + expr
	b.log.Warnf(expr2, vars...)
}
