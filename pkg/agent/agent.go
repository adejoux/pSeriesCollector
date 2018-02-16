package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/bus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/devices/hmc"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/agent/selfmon"
	"github.com/adejoux/pSeriesCollector/pkg/config"
)

// Version X.Y.Z based versioning
var (
	Version    string
	Commit     string
	Branch     string
	BuildStamp string
)

// RInfo  Release basic version info for the agent
type RInfo struct {
	InstanceID string
	Version    string
	Commit     string
	Branch     string
	BuildStamp string
}

// GetRInfo return Release Agent Information
func GetRInfo() *RInfo {
	info := &RInfo{
		InstanceID: MainConfig.General.InstanceID,
		Version:    Version,
		Commit:     Commit,
		Branch:     Branch,
		BuildStamp: BuildStamp,
	}
	return info
}

var (
	// Bus the bus messaging system to send messages over the hmc
	Bus = bus.NewBus()

	// MainConfig has all configuration
	MainConfig config.Config

	// DBConfig db config
	DBConfig config.DBConfig

	log *logrus.Logger
	//mutex for hmc map
	mutex sync.RWMutex
	//reload mutex
	reloadMutex   sync.Mutex
	reloadProcess bool
	//runtime hmc
	hmcServers map[string]*hmc.HMCServer
	//runtime output db's
	influxdb map[string]*output.InfluxDB

	selfmonProc *selfmon.SelfMon
	// for synchronize  deivce specific goroutines
	gatherWg sync.WaitGroup
	senderWg sync.WaitGroup
)

// SetLogger set log output
func SetLogger(l *logrus.Logger) {
	log = l
}

//Reload Mutex Related Methods.

// CheckReloadProcess check if the agent is doing a reloading just now
func CheckReloadProcess() bool {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	return reloadProcess
}

// CheckAndSetReloadProcess set the reloadProcess flat to true and  return the last stat before true set
func CheckAndSetReloadProcess() bool {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	retval := reloadProcess
	reloadProcess = true
	return retval
}

// CheckAndUnSetReloadProcess set the reloadProcess flat to false and  return the last stat before true set
func CheckAndUnSetReloadProcess() bool {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	retval := reloadProcess
	reloadProcess = false
	return retval
}

//PrepareInfluxDBs review all configured db's in the SQL database
// and check if exist at least a "default", if not creates a dummy db which does nothing
func PrepareInfluxDBs() map[string]*output.InfluxDB {
	idb := make(map[string]*output.InfluxDB)

	var defFound bool
	for k, c := range DBConfig.Influxdb {
		//Inticialize each Influx device
		if k == "default" {
			defFound = true
		}
		idb[k] = output.NewNotInitInfluxDB(c)
	}
	if defFound == false {
		//no devices configured  as default device we need to set some device as itcan send data transparent to HMC devices goroutines
		log.Warn("No Output default found influxdb devices found !!")
		idb["default"] = output.DummyDB
	}
	return idb
}

//GetDevice is a safe method to get a Device Object
func GetDevice(id string) (*hmc.HMCServer, error) {
	var dev *hmc.HMCServer
	var ok bool
	if CheckReloadProcess() == true {
		log.Warning("There is a reload process running while trying to get device info")
		return nil, fmt.Errorf("There is a reload process running.... please wait until finished ")
	}
	mutex.RLock()
	if dev, ok = hmcServers[id]; !ok {
		return nil, fmt.Errorf("there is not any device with id %s running", id)
	}
	mutex.RUnlock()
	return dev, nil
}

//GetDeviceJSONInfo get device data in JSON format just if not doing a reloading process
func GetDeviceJSONInfo(id string) ([]byte, error) {
	var dev *hmc.HMCServer
	var ok bool
	if CheckReloadProcess() == true {
		log.Warning("There is a reload process running while trying to get device info")
		return nil, fmt.Errorf("There is a reload process running.... please wait until finished ")
	}
	mutex.RLock()
	defer mutex.RUnlock()
	if dev, ok = hmcServers[id]; !ok {
		return nil, fmt.Errorf("there is not any device with id %s running", id)
	}
	return dev.ToJSON()
}

// GetDevStats xx
func GetDevStats() map[string]*hmc.DevStat {
	devstats := make(map[string]*hmc.DevStat)
	mutex.RLock()
	for k, v := range hmcServers {
		devstats[k] = v.GetBasicStats()
	}
	mutex.RUnlock()
	return devstats
}

// StopInfluxOut xx
func StopInfluxOut(idb map[string]*output.InfluxDB) {
	for k, v := range idb {
		log.Infof("Stopping Influxdb out %s", k)
		v.StopSender()
	}
}

// ReleaseInfluxOut xx
func ReleaseInfluxOut(idb map[string]*output.InfluxDB) {
	for k, v := range idb {
		log.Infof("Release Influxdb resources %s", k)
		v.End()
	}
}

// DeviceProcessStop stop all device goroutines
func DeviceProcessStop() {
	Bus.Broadcast(&bus.Message{Type: "exit"})
}

// DeviceProcessStart start all hmc goroutines
func DeviceProcessStart() {
	mutex.RLock()

	for s, c := range hmcServers {
		log.Infof("Starting HMC device %s", s)
		c.StartGather(&gatherWg)
	}
	mutex.RUnlock()
}

// ReleaseDevices Executes End for each device
func ReleaseDevices() {
	mutex.RLock()
	for _, c := range hmcServers {
		c.End()
	}
	mutex.RUnlock()
}

func init() {
	go Bus.Start()
}

func initSelfMonitoring(idb map[string]*output.InfluxDB) {
	log.Debugf("INFLUXDB2: %+v", idb)
	selfmonProc = selfmon.NewNotInit(&MainConfig.Selfmon)

	if MainConfig.Selfmon.Enabled {
		if val, ok := idb["default"]; ok {
			//only executed if a "default" influxdb exist
			val.Init()
			val.StartSender(&senderWg)

			selfmonProc.Init()
			selfmonProc.SetOutDB(idb)
			selfmonProc.SetOutput(val)

			log.Printf("SELFMON enabled %+v", MainConfig.Selfmon)
			//Begin the statistic reporting
			selfmonProc.StartGather(&gatherWg)
		} else {
			MainConfig.Selfmon.Enabled = false
			log.Errorf("SELFMON disabled becaouse of no default db found !!! SELFMON[ %+v ]  INFLUXLIST[ %+v]\n", MainConfig.Selfmon, idb)
		}
	} else {
		log.Printf("SELFMON disabled %+v", MainConfig.Selfmon)
	}
}

// LoadConf call to initialize alln configurations
func LoadConf() {
	//Load all database info to Cfg struct
	MainConfig.Database.LoadDbConfig(&DBConfig)
	//Prepare the InfluxDataBases Configuration
	influxdb = PrepareInfluxDBs()

	// beginning self monitoring process if needed.( before each other gorotines could begin)

	initSelfMonitoring(influxdb)

	//Initialize Device Metrics CFG

	config.Init(&DBConfig)

	//Initialize Device Runtime map
	mutex.Lock()
	hmcServers = make(map[string]*hmc.HMCServer)
	mutex.Unlock()

	for k, c := range DBConfig.HMC {
		//Inticialize each HMC device and put pointer to the global map hmc
		dev := hmc.New(c)
		dev.AttachToBus(Bus)
		dev.SetSelfMonitoring(selfmonProc)
		//send db's map to initialize each one its own db if needed and not yet initialized

		outdb, _ := dev.GetOutSenderFromMap(influxdb)
		outdb.Init()
		outdb.StartSender(&senderWg)

		mutex.Lock()
		hmcServers[k] = dev
		mutex.Unlock()
	}

	//beginning  the gather process
}

// ReloadConf call to reinitialize alln configurations
func ReloadConf() (time.Duration, error) {
	start := time.Now()
	if CheckAndSetReloadProcess() == true {
		log.Warning("RELOADCONF: There is another reload process running while trying to reload at %s  ", start.String())
		return time.Since(start), fmt.Errorf("There is another reload process running.... please wait until finished ")
	}

	log.Infof("RELOADCONF INIT: begin HMC Gather processes stop... at %s", start.String())
	//stop all device prcesses
	DeviceProcessStop()
	log.Info("RELOADCONF: begin selfmon Gather processes stop...")
	//stop the selfmon process
	selfmonProc.StopGather()
	log.Info("RELOADCONF: waiting for all Gather gorotines stop...")
	//wait until Done
	gatherWg.Wait()
	log.Info("RELOADCONF: releasing Device Resources")
	ReleaseDevices()
	log.Info("RELOADCONF: releasing Seflmonitoring Resources")
	selfmonProc.End()
	log.Info("RELOADCONF: begin sender processes stop...")
	//stop all Output Emmiter
	//log.Info("DEBUG Gather WAIT %+v", GatherWg)
	//log.Info("DEBUG SENDER WAIT %+v", senderWg)
	StopInfluxOut(influxdb)
	log.Info("RELOADCONF: waiting for all Sender gorotines stop..")
	senderWg.Wait()
	log.Info("RELOADCONF: releasing Sender Resources")
	ReleaseInfluxOut(influxdb)

	log.Info("RELOADCONF: ĺoading configuration Again...")
	LoadConf()
	log.Info("RELOADCONF: Starting all device processes again...")
	DeviceProcessStart()
	log.Infof("RELOADCONF END: Finished from %s to %s [Duration : %s]", start.String(), time.Now().String(), time.Since(start).String())
	CheckAndUnSetReloadProcess()

	return time.Since(start), nil
}
