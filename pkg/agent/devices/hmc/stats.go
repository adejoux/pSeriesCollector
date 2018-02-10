package hmc

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/selfmon"
)

// DevStatType a device stat type
type DevStatType uint

const (
	// MetricSent all values had been sent (measurment fields -- could be from OID's or from computed, evaluated, sources)
	MetricSent = 1
	// MetricSentErrors values that has errors when trying to add to a measurement
	MetricSentErrors = 2
	// MeasurementSent all measurements sent to the influx backend
	MeasurementSent = 3
	// MeasurementSentErrors all measurements with errors
	MeasurementSentErrors = 4
	// CycleGatherStartTime Time which begins the last Gather Cycle
	CycleGatherStartTime = 5
	// CycleGatherDuration Time taken in complete the last gather and sent cycle
	CycleGatherDuration = 6
	// BackEndSentStartTime Time witch begins the last sent process
	BackEndSentStartTime = 7
	// BackEndSentDuration Time taken in complete the data sent process
	BackEndSentDuration = 8
	// DevStatTypeSize special value to set the last stat position
	DevStatTypeSize = 9
)

// DevStat minimal info to show users
type DevStat struct {
	//ID
	id     string
	TagMap map[string]string
	//Control
	log     *logrus.Logger
	selfmon *selfmon.SelfMon
	mutex   sync.Mutex

	//Counter Statistics
	Counters []interface{}

	//device state
	DeviceActive    bool
	DeviceConnected bool
	//extra measurement statistics
	NumMetrics int
}

// Init initializes the device stat object
func (s *DevStat) Init(id string, tm map[string]string, l *logrus.Logger) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.id = id
	s.TagMap = tm
	s.log = l
	s.Counters = make([]interface{}, DevStatTypeSize)
	s.Counters[MetricSent] = 0
	s.Counters[MeasurementSent] = 0
	s.Counters[MetricSentErrors] = 0
	s.Counters[MeasurementSentErrors] = 0
	s.Counters[CycleGatherStartTime] = 0
	s.Counters[CycleGatherDuration] = 0.0
	s.Counters[BackEndSentStartTime] = 0
	s.Counters[BackEndSentDuration] = 0.0
}

func (s *DevStat) reset() {
	for k, val := range s.Counters {
		switch v := val.(type) {
		case string:
			s.Counters[k] = ""
		case int32, int64, int:
			s.Counters[k] = 0
		case float64, float32:
			s.Counters[k] = 0.0
		default:
			s.log.Warnf("unknown typpe for counter %#v", v)
		}
	}
}

// GetCounter get Counter for stats
func (s *DevStat) GetCounter(stat DevStatType) interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.Counters[stat]
}

func (s *DevStat) getMetricFields() map[string]interface{} {
	fields := map[string]interface{}{
		/*1*/ "metric_sent": s.Counters[MetricSent],
		/*2*/ "metric_sent_errors": s.Counters[MetricSentErrors],
		/*3*/ "measurement_sent": s.Counters[MeasurementSent],
		/*4*/ "measurement_sent_errors": s.Counters[MeasurementSentErrors],
		/*5*/ "cycle_gather_start_time": s.Counters[CycleGatherStartTime],
		/*6*/ "cycle_gather_duration": s.Counters[CycleGatherDuration],
		/*7*/ "backend_sent_start_time": s.Counters[BackEndSentStartTime],
		/*8*/ "backend_sent_duration": s.Counters[BackEndSentDuration],
	}
	return fields
}

// SetSelfMonitoring set the output device where send monitoring metrics
func (s *DevStat) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.selfmon = cfg
}

// ThSafeCopy get a new object with public data copied in thread safe way
func (s *DevStat) ThSafeCopy() *DevStat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	st := &DevStat{}
	st.Init(s.id, s.TagMap, s.log)
	for k, v := range s.Counters {
		st.Counters[k] = v
	}
	return st
}

// Send send data to the selfmon device
func (s *DevStat) Send() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.log.Infof("STATS HMC :  polling took [%f seconds] ", s.Counters[CycleGatherDuration])
	s.log.Infof("STATS INFLUX: influx send took [%f seconds]", s.Counters[BackEndSentDuration])
	if s.selfmon != nil {
		s.selfmon.AddDeviceMetrics(s.id, s.getMetricFields(), s.TagMap)
	}
}

// ResetCounters initialize metric counters
func (s *DevStat) ResetCounters() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.reset()
}

// CounterInc n values to the counter set by id
func (s *DevStat) CounterInc(id DevStatType, n int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[id] = s.Counters[id].(int) + int(n)
}

// AddMeasStats add measurement stats to the device stats object
func (s *DevStat) AddMeasStats(mets int64, mete int64, meass int64, mease int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[MetricSent] = s.Counters[MetricSent].(int) + int(mets)
	s.Counters[MetricSentErrors] = s.Counters[MetricSentErrors].(int) + int(mete)
	s.Counters[MeasurementSent] = s.Counters[MeasurementSent].(int) + int(meass)
	s.Counters[MeasurementSentErrors] = s.Counters[MeasurementSentErrors].(int) + int(mease)
}

// SetGatherDuration Update Gather Duration stats
func (s *DevStat) SetGatherDuration(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[CycleGatherStartTime] = start.Unix()
	s.Counters[CycleGatherDuration] = duration.Seconds()
}

// AddSentDuration Update Sent Duration stats
func (s *DevStat) AddSentDuration(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//only register the first start time on concurrent mode
	if s.Counters[BackEndSentStartTime] == 0 {
		s.Counters[BackEndSentStartTime] = start.Unix()
	}
	s.Counters[BackEndSentDuration] = s.Counters[BackEndSentDuration].(float64) + duration.Seconds()
}
