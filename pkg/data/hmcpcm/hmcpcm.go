package hmcpcm

import (
	"github.com/Sirupsen/logrus"
)

var (
	log *logrus.Logger
)

//mutex for devices m

// SetLogger set log output
func SetLogger(l *logrus.Logger) {
	log = l
}
