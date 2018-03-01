package config

import (
	"fmt"
	"strings"
	// _ needed to mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	"os"
	"sync/atomic"
	// _ needed to sqlite3
	_ "github.com/mattn/go-sqlite3"
)

func (dbc *DatabaseCfg) resetChanges() {
	atomic.StoreInt64(&dbc.numChanges, 0)
}

func (dbc *DatabaseCfg) addChanges(n int64) {
	atomic.AddInt64(&dbc.numChanges, n)
}
func (dbc *DatabaseCfg) getChanges() int64 {
	return atomic.LoadInt64(&dbc.numChanges)
}

//DbObjAction measurement groups to assign to devices
type DbObjAction struct {
	Type     string
	TypeDesc string
	ObID     string
	Action   string
}

//InitDB initialize de BD configuration
func (dbc *DatabaseCfg) InitDB() {
	// Create ORM engine and database
	var err error
	var dbtype string
	var datasource string

	log.Debugf("Database config: %+v", dbc)

	switch dbc.Type {
	case "sqlite3":
		dbtype = "sqlite3"
		datasource = dataDir + "/" + dbc.Name + ".db"
	case "mysql":
		dbtype = "mysql"
		protocol := "tcp"
		if strings.HasPrefix(dbc.Host, "/") {
			protocol = "unix"
		}
		datasource = fmt.Sprintf("%s:%s@%s(%s)/%s?charset=utf8", dbc.User, dbc.Password, protocol, dbc.Host, dbc.Name)

	default:
		log.Errorf("unknown db  type %s", dbc.Type)
		return
	}

	dbc.x, err = xorm.NewEngine(dbtype, datasource)
	if err != nil {
		log.Fatalf("Fail to create engine: %v\n", err)
	}

	if len(dbc.SQLLogFile) != 0 {
		dbc.x.ShowSQL(true)
		f, error := os.Create(logDir + "/" + dbc.SQLLogFile)
		if err != nil {
			log.Errorln("Fail to create log file  ", error)
		}
		dbc.x.SetLogger(xorm.NewSimpleLogger(f))
	}
	if dbc.Debug == "true" {
		dbc.x.Logger().SetLevel(core.LOG_DEBUG)
	}

	// Sync tables
	if err = dbc.x.Sync(new(InfluxCfg)); err != nil {
		log.Fatalf("Fail to sync database InfluxServerCfg: %v\n", err)
	}

	if err = dbc.x.Sync(new(HMCCfg)); err != nil {
		log.Fatalf("Fail to sync database HMCConfig: %v\n", err)
	}

	if err = dbc.x.Sync(new(DeviceCfg)); err != nil {
		log.Fatalf("Fail to sync database DeviceCfg: %v\n", err)
	}
}

//LoadDbConfig get data from database
func (dbc *DatabaseCfg) LoadDbConfig(cfg *DBConfig) {
	var err error

	//Load InfluxDB engines map
	cfg.Influxdb, err = dbc.GetInfluxCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Influx Ouput servers URL :%v", err)
	}

	//Load HMC engines map
	cfg.HMC, err = dbc.GetHMCCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Influx Ouput servers URL :%v", err)
	}

	//Load Devices map
	cfg.Devices, err = dbc.GetDeviceCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Influx Ouput servers URL :%v", err)
	}

	dbc.resetChanges()
}
