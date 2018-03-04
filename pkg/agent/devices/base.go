package devices

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/bus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/agent/selfmon"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
	"os"
	"sync"
	"time"
)

// Base the Base strut to do a gather procesing
type Base struct {
	ID   string
	Type string
	//log
	logPrefix   string
	log         *logrus.Logger
	CurLogLevel string
	//TagMap
	TagMap map[string]string
	//Scan
	UpdateScanFreq int
	//stats
	RtStats DevStat          //Runtime Internal statistic
	Stats   *DevStat         //Public info for thread safe accessing to the data ()
	Influx  *output.InfluxDB `json:"-"`
	Freq    int
	//runtime controls
	statsData sync.RWMutex
	rtData    sync.RWMutex
	//device status
	DeviceActive       bool
	DeviceConnected    bool
	StateDebug         bool
	Node               *bus.Node `json:"-"`
	ReloadLoopsPending int
	//custom   Methods
	Gather                  func()                    `json:"-"`
	Scan                    func() error              `json:"-"`
	ReleaseClient           func()                    `json:"-"`
	Reconnect               func() error              `json:"-"`
	CheckDeviceConnectivity func() bool               `json:"-"`
	SetProtocolDebug        func(bool)                `json:"-"`
	HandleMessages          func(string, interface{}) `json:"-"`
}

//-------------------------------
// INIT Object Handlers
//-------------------------------

// Init return logger
func (b *Base) Init(object interface{}, id string) {
	b.ID = id
	b.Type = fmt.Sprintf("%T", object)
	b.UpdateScanFreq = 60
}

// SetScanFreq set frequency
func (b *Base) SetScanFreq(fr int) {
	b.UpdateScanFreq = fr
	b.setReloadLoopsPending(fr)
}

//-------------------------------
// Data Lock Handlers
//-------------------------------

// DataLock lock data if needed
func (b *Base) DataLock() {
	b.rtData.RLock()
}

// DataUnlock data if needed
func (b *Base) DataUnlock() {
	b.rtData.RUnlock()
}

//-------------------------------
// LOG Handlers
//-------------------------------

// GetLogger return logger
func (b *Base) GetLogger() *logrus.Logger {
	return b.log
}

// InitLog set the log file
func (b *Base) InitLog(LogFile string, LogLevel string) {
	b.logPrefix = "[" + b.Type + "] [" + b.ID + "]"
	f, _ := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	b.log = logrus.New()
	b.log.Out = f
	l, _ := logrus.ParseLevel(LogLevel)
	b.log.Level = l
	b.CurLogLevel = b.log.Level.String()
	//Formatter for time
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	b.log.Formatter = customFormatter
	customFormatter.FullTimestamp = true
}

//-------------------------------
// Bus Handlers
//-------------------------------

// AttachToBus add this device to a communition bus
func (b *Base) AttachToBus(id string, B *bus.Bus) {
	b.Node = bus.NewNode(id)
	B.Join(b.Node)
}

//-------------------------------
// Stats Handlers
//-------------------------------

// InitStats initialize the base statistics
func (b *Base) InitStats(id string) {
	b.RtStats.Init(id, b.TagMap, b.log)
	b.Stats = b.GetBaseStats()
}

// SetSelfMonitoring set the output device where send monitoring metrics
func (b *Base) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	b.RtStats.SetSelfMonitoring(cfg)
}

// GetBasicStats get basic info for this device
func (b *Base) GetBasicStats() *DevStat {
	return b.Stats
}

// GetBaseStats get basic info for this device
func (b *Base) GetBaseStats() *DevStat {
	b.statsData.Lock()
	defer b.statsData.Unlock()
	sum := 0

	stat := b.RtStats.ThSafeCopy()
	stat.TagMap = b.TagMap
	stat.DeviceActive = b.DeviceActive
	stat.DeviceConnected = b.DeviceConnected
	stat.ReloadLoopsPending = b.ReloadLoopsPending

	stat.NumMetrics = sum

	return stat
}

//-------------------------------
// ReloadLoop Handlers
//-------------------------------

func (b *Base) setReloadLoopsPending(val int) {
	b.ReloadLoopsPending = val
}

func (b *Base) getReloadLoopsPending() int {
	return b.ReloadLoopsPending
}

func (b *Base) decReloadLoopsPending() {
	if b.ReloadLoopsPending > 0 {
		b.ReloadLoopsPending--
	}
}

//------------------------------
// Bus messages
//------------------------------

// ForceGather send message to force a data gather execution
func (b *Base) ForceGather() {
	b.Node.SendMsg(&bus.Message{Type: "forcegather"})
}

// ForceDevScan send info to update the filter counter to the next execution
func (b *Base) ForceDevScan() {
	b.Node.SendMsg(&bus.Message{Type: "devscan"})
}

// StopGather send signal to stop the Gathering process
func (b *Base) StopGather() {
	b.Node.SendMsg(&bus.Message{Type: "exit"})
}

//RTActivate change activatio state in runtime
func (b *Base) RTActivate(activate bool) {
	b.Node.SendMsg(&bus.Message{Type: "enabled", Data: activate})
}

// RTProtocolDebug change HMC Session Debug runtime
func (b *Base) RTProtocolDebug(activate bool) {
	b.Node.SendMsg(&bus.Message{Type: "protocoldebug", Data: activate})
}

// RTSetLogLevel set the log level for this device
func (b *Base) RTSetLogLevel(level string) {
	b.Node.SendMsg(&bus.Message{Type: "loglevel", Data: level})
}

//-------------------------------
// Gathering Handlers
//-------------------------------

func (b *Base) gatherAndProcessData(t *time.Ticker, force bool) *time.Ticker {
	b.rtData.Lock()
	defer b.rtData.Unlock()
	//if active
	if b.DeviceActive || force {
	FORCEINIT:
		//check if device is connected
		if b.DeviceConnected == false {
			//should release first previous HMC connections
			b.ReleaseClient()
			//try reconnect
			err := b.Reconnect()
			if err == nil {
				b.DeviceConnected = true
				start := time.Now()
				b.Scan() //
				elapsed := time.Since(start)
				b.RtStats.SetScanStats(start, elapsed)
				b.Infof("SCAN finished in: %s", elapsed.String())

				if force == false {
					// Round collection to nearest interval by sleeping
					//and reprogram the ticker to aligned starts
					// only when no extra gather(forced from web-ui)
					utils.WaitAlignForNextCycle(b.Freq, b.GetLogger())
					t.Stop()
					t = time.NewTicker(time.Duration(b.Freq) * time.Second)
					//force one iteration now..after device has been connected  dont wait for next
					//ticker (1 complete cycle)
				}
				goto FORCEINIT
			}
		} else {
			//device active and connected
			b.Infof("Init gather cycle for device %s", b.ID)
			/*************************
			 *
			 * Device Gather data process
			 *
			 ***************************/
			b.RtStats.ResetCounters()
			b.Gather()

			/*******************************************
			 *
			 * ReScan Devices (if needed)
			 *
			 *******************************************/
			//Check if reload needed with b.ReloadLoopsPending if a posivive value on negative this will disabled

			b.decReloadLoopsPending()

			if b.getReloadLoopsPending() == 0 {
				start := time.Now()
				b.Scan()
				elapsed := time.Since(start)
				b.RtStats.SetScanStats(start, elapsed)
				b.Infof("SCAN finished in: %s", elapsed.String())
				b.setReloadLoopsPending(b.UpdateScanFreq)
			}

			b.DeviceConnected = b.CheckDeviceConnectivity()

			b.RtStats.Send()
		}
	} else {
		b.Infof("Gather process is disabled")
	}
	//get Ready a copy of the stats to

	b.Stats = b.GetBaseStats()

	return t
}

// StartGather Main GoRutine method to begin HMC data collecting
func (b *Base) StartGather(wg *sync.WaitGroup) {
	wg.Add(1)
	go b.startGatherGo(wg)
}

func (b *Base) startGatherGo(wg *sync.WaitGroup) {
	defer wg.Done()

	b.Infof("Starting Gathering goroutine...")

	if b.DeviceActive && b.DeviceConnected {
		b.Infof("Begin first InidevInfo")
		// REVIEW perhaps not needed here
		b.rtData.Lock()
		start := time.Now()
		b.Scan()
		elapsed := time.Since(start)
		b.RtStats.SetScanStats(start, elapsed)
		b.Infof("SCAN finished in: %s", elapsed.String())
		b.rtData.Unlock()
	} else {
		b.Infof("Can not initialize this device: Is Active: %t  |  Connection Active: %t ", b.DeviceActive, b.DeviceConnected)
	}

	t := time.NewTicker(time.Duration(b.Freq) * time.Second)
	for {

		t = b.gatherAndProcessData(t, false)

	LOOP:
		for {
			select {
			case <-t.C:
				break LOOP
			case val := <-b.Node.Read:
				b.Infof("Received Message...%s: %+v", val.Type, val.Data)
				switch val.Type {

				case "forcegather":
					b.Infof("invoked Force Data Gather And Process")
					b.gatherAndProcessData(t, true)
				case "exit":
					b.Infof("invoked EXIT from HMC Gather process ")
					return
				case "devscan":
					b.rtData.Lock()
					b.setReloadLoopsPending(1)
					b.rtData.Unlock()
				case "protocoldebug":
					debug := val.Data.(bool)
					if b.DeviceConnected == true {
						b.rtData.Lock()
						b.StateDebug = debug
						b.SetProtocolDebug(debug)
						b.rtData.Unlock()
					} else {
						b.Warnf("Device not connected we can not set debug yet")
					}

				case "enabled":
					status := val.Data.(bool)
					b.rtData.Lock()
					b.DeviceActive = status
					b.Infof("device STATUS  ACTIVE  [%t] ", status)
					b.rtData.Unlock()
				case "loglevel":
					level := val.Data.(string)
					l, err := logrus.ParseLevel(level)
					if err != nil {
						b.Warnf("ERROR on Changing LOGLEVEL to [%t] ", level)
						break
					}
					b.rtData.Lock()
					b.log.Level = l
					b.Infof("device loglevel Changed  [%s] ", level)
					b.CurLogLevel = b.log.Level.String()
					b.rtData.Unlock()
				default:
					b.HandleMessages(val.Type, val.Data)

				}
			}
			//Some online actions can change Stats

			b.Stats = b.GetBaseStats()

		}
	}
}
