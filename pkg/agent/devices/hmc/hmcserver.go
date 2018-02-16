package hmc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/bus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/agent/selfmon"
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/adejoux/pSeriesCollector/pkg/data/hmcpcm"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
)

var (
	cfg    *config.DBConfig
	logDir string
)

// SetDBConfig set agent config
func SetDBConfig(c *config.DBConfig) {
	cfg = c
}

// SetLogDir set log dir
func SetLogDir(l string) {
	logDir = l
}

// HMCServer contains all runtime device related device configu ns and state
type HMCServer struct {
	Session *hmcpcm.Session `json:"-"`
	cfg     *config.HMCCfg
	log     *logrus.Logger
	//runtime built TagMap
	TagMap map[string]string
	//Refresh data to show in the frontend
	Freq int

	Influx *output.InfluxDB `json:"-"`
	//LastError     time.Time
	//Runtime stats
	stats DevStat  //Runtime Internal statistic
	Stats *DevStat //Public info for thread safe accessing to the data ()

	//runtime controls
	rtData    sync.RWMutex
	statsData sync.RWMutex

	DeviceActive    bool
	DeviceConnected bool
	StateDebug      bool

	Node *bus.Node `json:"-"`

	CurLogLevel string
	Gather      func() `json:"-"`

	ManagedSystemOnly bool
}

// Ping get data from hmc
func Ping(c *config.HMCCfg, log *logrus.Logger, apidbg bool, filename string) (*hmcpcm.Session, time.Duration, string, error) {
	HMCURL := fmt.Sprintf("https://"+"%s"+":12443", c.Host)
	start := time.Now()

	Session, err := hmcpcm.NewSession(HMCURL, c.User, c.Password)
	if err != nil {
		return nil, 0, "error", err
	}
	if log != nil {
		Session.SetLog(log)
	}
	if apidbg == true {
		Session.Debug = true
		Session.SetDebugLog(filename)
	}

	err = Session.DoLogon()
	if err != nil {
		return nil, 0, "error", err
	}
	elapsed := time.Since(start)
	return Session, elapsed, "test", nil
}

// New create and Initialice a device Object
func New(c *config.HMCCfg) *HMCServer {
	dev := HMCServer{}
	dev.Init(c)
	return &dev
}

// GetLogFilePath return current LogFile
func (d *HMCServer) GetLogFilePath() string {
	return d.cfg.LogFile
}

// ToJSON return a JSON version of the device data
func (d *HMCServer) ToJSON() ([]byte, error) {

	d.rtData.RLock()
	defer d.rtData.RUnlock()
	result, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		d.Errorf("Error on Get JSON data from device")
		dummy := []byte{}
		return dummy, nil
	}
	return result, err
}

// GetBasicStats get basic info for this device
func (d *HMCServer) GetBasicStats() *DevStat {
	d.statsData.RLock()
	defer d.statsData.RUnlock()
	return d.Stats
}

// GetBasicStats get basic info for this device
func (d *HMCServer) getBasicStats() *DevStat {

	sum := 0

	stat := d.stats.ThSafeCopy()
	stat.TagMap = d.TagMap
	stat.DeviceActive = d.DeviceActive
	stat.DeviceConnected = d.DeviceConnected

	stat.NumMetrics = sum

	return stat
}

// GetOutSenderFromMap to get info about the sender will use
func (d *HMCServer) GetOutSenderFromMap(influxdb map[string]*output.InfluxDB) (*output.InfluxDB, error) {
	if len(d.cfg.OutDB) == 0 {
		d.Warnf("No OutDB configured on the device")
	}
	var ok bool
	name := d.cfg.OutDB
	if d.Influx, ok = influxdb[name]; !ok {
		//we assume there is always a default db
		if d.Influx, ok = influxdb["default"]; !ok {
			//but
			return nil, fmt.Errorf("No influx config for HMC device: %s", d.cfg.ID)
		}
	}

	return d.Influx, nil
}

// ForceGather send message to force a data gather execution
func (d *HMCServer) ForceGather() {
	d.Node.SendMsg(&bus.Message{Type: "forcegather"})
}

// StopGather send signal to stop the Gathering process
func (d *HMCServer) StopGather() {
	d.Node.SendMsg(&bus.Message{Type: "exit"})
}

//RTActivate change activatio state in runtime
func (d *HMCServer) RTActivate(activate bool) {
	d.Node.SendMsg(&bus.Message{Type: "enabled", Data: activate})
}

//RTActHMCAPIDebug change HMC Session Debug runtime
func (d *HMCServer) RTActHMCAPIDebug(activate bool) {
	d.Node.SendMsg(&bus.Message{Type: "hmcapidebug", Data: activate})
}

// RTSetLogLevel set the log level for this device
func (d *HMCServer) RTSetLogLevel(level string) {
	d.Node.SendMsg(&bus.Message{Type: "loglevel", Data: level})
}

//InitDevMeasurements generte all meeded internal structs
func (d *HMCServer) InitDevMeasurements() {

}

// this method puts all metrics as invalid once sent to the backend
// it lets us to know if any of them has not been updated in the gathering process
func (d *HMCServer) invalidateMetrics() {
	/*	for _, v := range d.Measurements {
		v.InvalidateMetrics()
	}*/
}

// GetHMCData get data from HMC
func (d *HMCServer) GetHMCData() {

	bpts, _ := d.Influx.BP()
	startStats := time.Now()

	points := NewPointArray(d.log, bpts)
	//prepare batchpoint
	err := d.ImportData(points)
	if err != nil {
		d.Errorf("Error in  import Data to HMC %s: ERROR: %s", d.cfg.ID, err)
		return
	}
	d.stats.AddMeasStats(points.MetSent, points.MetError, points.MeasSent, points.MeasError)
	points.Flush()
	elapsedStats := time.Since(startStats)

	d.stats.SetGatherDuration(startStats, elapsedStats)
	/*************************
	 *
	 * Send data to InfluxDB process
	 *
	 ***************************/

	startInfluxStats := time.Now()
	if bpts != nil {
		d.Influx.Send(bpts)
	} else {
		d.Warnf("Can not send data to the output DB becaouse of batchpoint creation error")
	}
	elapsedInfluxStats := time.Since(startInfluxStats)
	d.stats.AddSentDuration(startInfluxStats, elapsedInfluxStats)
}

/*
Init  does the following

- Initialize not set variables to some defaults
- Initialize logfile for this device
- Initialize comunication channels and initial device state
*/
func (d *HMCServer) Init(c *config.HMCCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	d.cfg = c

	//Init Freq
	d.Freq = d.cfg.Freq
	if d.cfg.Freq == 0 {
		d.Freq = 60
	}

	if d.cfg.Freq < 30 {
		d.Freq = 30
		d.Warnf("Error en the gathering frequency got (%d) , but forced  to 30 seconds as the min HMC resolution", d.cfg.Freq)
	}

	//Init Logger
	if len(d.cfg.LogFile) == 0 {
		d.cfg.LogFile = logDir + "/" + d.cfg.ID + ".log"

	}
	if len(d.cfg.LogLevel) == 0 {
		d.cfg.LogLevel = "info"
	}

	f, _ := os.OpenFile(d.cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	d.log = logrus.New()
	d.log.Out = f
	l, _ := logrus.ParseLevel(d.cfg.LogLevel)
	d.log.Level = l
	d.CurLogLevel = d.log.Level.String()
	//Formatter for time
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	d.log.Formatter = customFormatter
	customFormatter.FullTimestamp = true

	d.ManagedSystemOnly = d.cfg.ManagedSystemsOnly

	d.DeviceActive = d.cfg.Active

	//Init Device Tags

	d.TagMap = make(map[string]string)
	if len(d.cfg.DeviceTagName) == 0 {
		d.cfg.DeviceTagName = "hmc"
	}

	conerr := d.Reconnect()
	if conerr != nil {
		d.Errorf("First HMC connect error: %s", conerr)
		d.DeviceConnected = false
	} else {
		d.DeviceConnected = true
	}

	var val string

	switch d.cfg.DeviceTagValue {
	case "id":
		val = d.cfg.ID
	case "host":
		val = d.cfg.Host
	default:
		val = d.cfg.ID
		d.Warnf("Unkwnown DeviceTagValue %s set ID (%s) as value", d.cfg.DeviceTagValue, val)
	}

	d.TagMap[d.cfg.DeviceTagName] = val

	if len(d.cfg.ExtraTags) > 0 {
		for _, tag := range d.cfg.ExtraTags {
			s := strings.Split(tag, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				d.TagMap[key] = value
			} else {
				d.Errorf("Error on tag definition TAG=VALUE [ %s ]", tag)
			}
		}
	} else {
		d.Warnf("No map detected in device")
	}
	// Init stats
	d.stats.Init(d.cfg.ID, d.TagMap, d.log)

	d.Gather = d.GetHMCData

	d.statsData.Lock()
	d.Stats = d.getBasicStats()
	d.statsData.Unlock()
	return nil
}

// AttachToBus add this device to a communition bus
func (d *HMCServer) AttachToBus(b *bus.Bus) {
	d.Node = bus.NewNode(d.cfg.ID)
	b.Join(d.Node)
}

// End The Opposite of Init() uninitialize all variables
func (d *HMCServer) End() {
	d.Node.Close()
	d.Session.Release()
}

// ReleaseClient release connections
func (d *HMCServer) ReleaseClient() {
	d.Session.Release()
}

// Reconnect does HTTP connection  protocol
func (d *HMCServer) Reconnect() error {
	var t time.Duration
	var id string
	var err error
	d.Debugf("Trying Reconnect again....")
	d.Session, t, id, err = Ping(d.cfg, d.log, d.cfg.HMCAPIDebug, d.cfg.ID)
	if err != nil {
		d.Errorf("Error on HMC connection %s", err)
		return err
	}
	//We should set samples to get only for the latest period
	d.Session.SetSamples((d.Freq / 30) + 1)

	d.Infof("Connected to HMC  OK : ID: %s : Duration %s ", id, t.String())
	return nil
}

// SetSelfMonitoring set the output device where send monitoring metrics
func (d *HMCServer) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	d.stats.SetSelfMonitoring(cfg)
}

// CheckDeviceConnectivity check if HMC connection is ok
func (d *HMCServer) CheckDeviceConnectivity() {
	/*
		ProcessedStat := d.stats.GetCounter(SnmpOIDGetProcessed)

		if value, ok := ProcessedStat.(int); ok {
			//check if no processed SNMP data (when this happens means there is not connectivity with the device )
			if value == 0 {
				d.DeviceConnected = false
			}
		} else {
			d.Warnf("Error in check Processd Stats %#+v ", ProcessedStat)
		}*/
}

func (d *HMCServer) gatherAndProcessData(t *time.Ticker, force bool) *time.Ticker {
	d.rtData.Lock()
	//if active
	if d.DeviceActive || force {
	FORCEINIT:
		//check if device is connected
		if d.DeviceConnected == false {
			//should release first previous HMC connections
			d.ReleaseClient()
			//try reconnect
			err := d.Reconnect()
			if err == nil {
				d.DeviceConnected = true
				//REVIEW perhaps not needed here
				d.InitDevMeasurements()

				if force == false {
					// Round collection to nearest interval by sleeping
					//and reprogram the ticker to aligned starts
					// only when no extra gather(forced from web-ui)
					utils.WaitAlignForNextCycle(d.Freq, d.log)
					t.Stop()
					t = time.NewTicker(time.Duration(d.Freq) * time.Second)
					//force one iteration now..after device has been connected  dont wait for next
					//ticker (1 complete cycle)
				}
				goto FORCEINIT
			}
		} else {
			//device active and connected
			d.Infof("Init gather cycle %s", d.cfg.ID)
			/*************************
			 *
			 * HMC Gather data process
			 *
			 ***************************/
			d.invalidateMetrics()
			d.stats.ResetCounters()
			d.Gather()

			d.CheckDeviceConnectivity()

			d.stats.Send()
		}
	} else {
		d.Infof("Gather process is disabled")
	}
	//get Ready a copy of the stats to

	d.statsData.Lock()
	d.Stats = d.getBasicStats()
	d.statsData.Unlock()
	d.rtData.Unlock()
	return t
}

// StartGather Main GoRutine method to begin HMC data collecting
func (d *HMCServer) StartGather(wg *sync.WaitGroup) {
	wg.Add(1)
	go d.startGatherGo(wg)
}

func (d *HMCServer) startGatherGo(wg *sync.WaitGroup) {
	defer wg.Done()

	d.Infof("Starting Gathering goroutine...")

	if d.DeviceActive && d.DeviceConnected {
		d.Infof("Begin first InidevInfo")
		// REVIEW perhaps not needed here
		d.rtData.Lock()
		d.InitDevMeasurements()
		d.rtData.Unlock()
	} else {
		d.Infof("Can not initialize this device: Is Active: %t  |  Connection Active: %t ", d.DeviceActive, d.DeviceConnected)
	}

	d.Infof("Beginning gather process for device on host (%s)", d.cfg.Host)

	t := time.NewTicker(time.Duration(d.Freq) * time.Second)
	for {

		t = d.gatherAndProcessData(t, false)

	LOOP:
		for {
			select {
			case <-t.C:
				break LOOP
			case val := <-d.Node.Read:
				d.Infof("Received Message...%s: %+v", val.Type, val.Data)
				switch val.Type {
				case "forcegather":
					d.Infof("invoked Force Data Gather And Process")
					d.gatherAndProcessData(t, true)
				case "exit":
					d.Infof("invoked EXIT from HMC Gather process ")
					return
				case "hmcapidebug":
					debug := val.Data.(bool)
					if d.DeviceConnected == true {
						d.rtData.Lock()
						d.StateDebug = debug
						d.Session.Debug = debug
						d.Session.SetDebugLog(d.cfg.ID)
						d.rtData.Unlock()
					} else {
						d.Warnf("Device not connected we can not set debug yet")
					}

				case "enabled":
					status := val.Data.(bool)
					d.rtData.Lock()
					d.DeviceActive = status
					d.Infof("device STATUS  ACTIVE  [%t] ", status)
					d.rtData.Unlock()
				case "loglevel":
					level := val.Data.(string)
					l, err := logrus.ParseLevel(level)
					if err != nil {
						d.Warnf("ERROR on Changing LOGLEVEL to [%t] ", level)
						break
					}
					d.rtData.Lock()
					d.log.Level = l
					d.Infof("device loglevel Changed  [%s] ", level)
					d.CurLogLevel = d.log.Level.String()
					d.rtData.Unlock()
				}
			}
			//Some online actions can change Stats
			d.statsData.Lock()
			d.Stats = d.getBasicStats()
			d.statsData.Unlock()
		}
	}
}
