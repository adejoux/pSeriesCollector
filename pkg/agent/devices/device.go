package devices

import (
	"github.com/adejoux/pSeriesCollector/pkg/agent/bus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/agent/selfmon"

	"sync"
)

// Device Interface
type Device interface {
	//for WebUI
	ForceGather()
	ForceDevScan()
	GetLogFilePath() string
	RTSetLogLevel(string)
	RTActivate(bool)
	RTProtocolDebug(bool)
	//For agent
	ToJSON() ([]byte, error)
	AttachToBus(string, *bus.Bus)
	SetSelfMonitoring(*selfmon.SelfMon)
	GetOutSenderFromMap(map[string]*output.InfluxDB) (*output.InfluxDB, error)
	GetBasicStats() *DevStat
	StopGather()
	StartGather(*sync.WaitGroup)
	End()
}
