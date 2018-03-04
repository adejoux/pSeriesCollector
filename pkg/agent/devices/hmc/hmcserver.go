package hmc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/devices"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/adejoux/pSeriesCollector/pkg/data/hmcpcm"
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
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

// HMCServer contains all runtime device related device configu ns and state
type HMCServer struct {
	devices.Base
	Session           *hmcpcm.Session `json:"-"`
	cfg               *config.HMCCfg
	ManagedSystemOnly bool
	System            map[string]*hmcpcm.ManagedSystem
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

//ScanHMC scan hmc
func ScanHMC(ses *hmcpcm.Session) (map[string]*hmcpcm.ManagedSystem, error) {
	ret := make(map[string]*hmcpcm.ManagedSystem)
	data, err := ses.GetManagedSystems()
	if err != nil {
		return ret, fmt.Errorf("ERROR on get Managed Systems: %s", err)
	}

	for _, entry := range data.Entries {
		ret[entry.ID] = &entry.Contents[0].System[0]
	}
	return ret, nil
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
	d.DataLock()
	defer d.DataUnlock()
	result, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		d.Errorf("Error on Get JSON data from device")
		dummy := []byte{}
		return dummy, nil
	}
	return result, err
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

func (d *HMCServer) handleMessages(id string, data interface{}) {
	return
}

func (d *HMCServer) setProtocolDebug(debug bool) {
	d.Session.Debug = debug
	d.Session.SetDebugLog(d.cfg.ID)
}

// GetHMCData get data from HMC
func (d *HMCServer) GetHMCData() {

	bpts, _ := d.Influx.BP()
	startStats := time.Now()

	points := pointarray.New(d.GetLogger(), bpts)
	//prepare batchpoint
	err := d.ImportData(points)
	if err != nil {
		d.Errorf("Error in  import Data to HMC %s: ERROR: %s", d.cfg.ID, err)
		return
	}
	points.Flush()
	elapsedStats := time.Since(startStats)

	d.RtStats.SetGatherDuration(startStats, elapsedStats)
	d.RtStats.AddMeasStats(points.MetSent, points.MetError, points.MeasSent, points.MeasError)

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
	d.RtStats.AddSentDuration(startInfluxStats, elapsedInfluxStats)
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
	// Set ALL methods IMPORTANT!!! (review if interface could be better here)
	d.Gather = d.GetHMCData
	d.Scan = d.ScanHMCDevices
	d.ReleaseClient = d.ReleaseHMCClient
	d.Reconnect = d.HMCReconnect
	d.SetProtocolDebug = d.setProtocolDebug
	d.HandleMessages = d.handleMessages
	d.CheckDeviceConnectivity = d.CheckHMCConnectivity

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
	d.Base.Init(d, d.cfg.ID)
	d.InitLog(d.cfg.LogFile, d.cfg.LogLevel)

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

	ExtraTags, err := utils.KeyValArrayToMap(d.cfg.ExtraTags)
	if err != nil {
		d.Warnf("Warning on Device  %s Tag gathering: %s", err)
	}
	utils.MapAdd(d.TagMap, ExtraTags)

	// Init stats
	d.InitStats(d.cfg.ID)

	d.SetScanFreq(d.cfg.UpdateScanFreq)

	return nil
}

// End The Opposite of Init() uninitialize all variables
func (d *HMCServer) End() {
	d.Node.Close()
	d.Session.Release()
}

// ReleaseClient release connections
func (d *HMCServer) ReleaseHMCClient() {
	d.Session.Release()
}

// HMCReconnect does HTTP connection  protocol
func (d *HMCServer) HMCReconnect() error {
	var t time.Duration
	var id string
	var err error
	d.Debugf("Trying Reconnect again....")
	d.Session, t, id, err = Ping(d.cfg, d.GetLogger(), d.cfg.HMCAPIDebug, d.cfg.ID)
	if err != nil {
		d.Errorf("Error on HMC connection %s", err)
		return err
	}
	//We should set samples to get only for the latest period
	//(+1 set as a workarround for some holes sometimes when gathering time + scan time > freq) REVIEW!!
	d.Session.SetSamples((d.Freq / 30) + 1)

	d.Infof("Connected to HMC  OK : ID: %s : Duration %s ", id, t.String())
	return nil
}

// CheckDeviceConnectivity check if HMC connection is ok
func (d *HMCServer) CheckHMCConnectivity() bool {
	d.Debugf("Check HMC Connectivity: Nothing to do in the HMCSERVER")
	/*
		ProcessedStat := d.RtStats.GetCounter(SnmpOIDGetProcessed)

		if value, ok := ProcessedStat.(int); ok {
			//check if no processed SNMP data (when this happens means there is not connectivity with the device )
			if value == 0 {
				d.DeviceConnected = false
			}
		} else {
			d.Warnf("Error in check Processd Stats %#+v ", ProcessedStat)
		}*/
	return true
}
