package nmon

import (
	"github.com/adejoux/pSeriesCollector/pkg/config"
)

var (
	cfg    *config.DBConfig
	db     *config.DatabaseCfg
	logDir string
)

// SetDBConfig set agent config
func SetDBConfig(c *config.DBConfig, d *config.DatabaseCfg) {
	cfg = c
	db = d
}

// SetLogDir set log dir
func SetLogDir(l string) {
	logDir = l
}
