package nmon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent/devices"
	"github.com/adejoux/pSeriesCollector/pkg/agent/output"
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
	"github.com/adejoux/pSeriesCollector/pkg/data/rfile"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Server contains all runtime device related device configu ns and state
type Server struct {
	devices.Base
	Timezone   string
	cfg        *config.DeviceCfg
	sftpclient *sftp.Client
	sshclient  *ssh.Client
	NmonFile   *NmonFile
	FilterMap  map[string]*regexp.Regexp
}

// Ping check connection to the
func Ping(c *config.DeviceCfg, log *logrus.Logger, apidbg bool, filename string) (*sftp.Client, *ssh.Client, time.Duration, string, error) {
	start := time.Now()
	sshhost := ""
	if len(c.NmonIP) == 0 {
		c.NmonIP = c.Name
	}
	sftp, ssh, err := rfile.InitSFTP(c.NmonIP, c.NmonSSHUser, c.NmonSSHKey)
	elapsed := time.Since(start)
	if err != nil {
		log.Errorf("Error en SSH connection (host: %s user:%s key: %s)  Error: %s", sshhost, c.NmonSSHUser, c.NmonSSHKey, err)
		return nil, nil, elapsed, "test", err
	}

	return sftp, ssh, elapsed, "SFTP test", nil
}

//ScanNmonDevice scan Device
func (d *Server) ScanNmonDevice() error {
	return nil
}

// New create and Initialice a device Object
func New(c *config.DeviceCfg, tz string) *Server {

	dev := Server{Timezone: tz}

	dev.Init(c)
	dev.Infof("New Device created:  %s | Timezone ( %s )", c.Name, tz)
	//Creating filters
	if len(c.NmonFilters) > 0 {
		dev.FilterMap = make(map[string]*regexp.Regexp)
		for _, f := range c.NmonFilters {
			if len(strings.TrimSpace(f)) == 0 {
				dev.Warnf("User Filter is void [%s]", f)
				continue
			}
			dev.Infof("Found NMON Filter: %s", f)
			reg, err := regexp.Compile(f)
			if err != nil {
				dev.Errorf("Regex Error on filter %s, Error:%s", f, err)
				continue
			}
			dev.FilterMap[f] = reg
		}
	}
	return &dev
}

// ToJSON return a JSON version of the device data
func (d *Server) ToJSON() ([]byte, error) {
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
func (d *Server) GetOutSenderFromMap(influxdb map[string]*output.InfluxDB) (*output.InfluxDB, error) {
	if len(d.cfg.NmonOutDB) == 0 {
		d.Warnf("GetOutSenderFromMap No OutDB configured on the device")
	}
	var ok bool
	name := d.cfg.NmonOutDB
	if d.Influx, ok = influxdb[name]; !ok {
		//we assume there is always a default db
		if d.Influx, ok = influxdb["default"]; !ok {
			//but
			return nil, fmt.Errorf("No influx config for the device: %s", d.cfg.ID)
		}
	}
	d.Debugf("GetOutSenderFromMap: This NMON server has configured the %s influxdb : %#+v", name, d.Influx)

	return d.Influx, nil
}

func (d *Server) handleMessages(id string, data interface{}) {

}

func (d *Server) setProtocolDebug(debug bool) {

}

// GetNmonData get data from Device
func (d *Server) GetNmonData() {

	bpts, _ := d.Influx.BP()
	startStats := time.Now()

	points := pointarray.New(d.GetLogger(), bpts)
	//prepare batchpoint
	err := d.ImportData(points)
	if err != nil {
		d.Errorf("Error in  import Nmon Data from Device %s: ERROR: %s", d.cfg.ID, err)
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
func (d *Server) Init(c *config.DeviceCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	// Set ALL methods IMPORTANT!!! (review if interface could be better here)
	d.Gather = d.GetNmonData
	d.Scan = d.ScanNmonDevice
	d.ReleaseClient = d.releaseClient
	d.Reconnect = d.reconnect
	d.SetProtocolDebug = d.setProtocolDebug
	d.HandleMessages = d.handleMessages
	d.CheckDeviceConnectivity = d.checkDeviceConnectivity

	d.cfg = c

	//Init Freq
	d.Freq = d.cfg.NmonFreq
	if d.cfg.NmonFreq == 0 {
		d.Freq = 60
	}

	//Init Logger

	d.Base.Init(d, d.cfg.Name)
	d.InitLog(logDir+"/"+d.cfg.Name+".log", d.cfg.NmonLogLevel)

	d.DeviceActive = d.cfg.EnableNmonStats

	//Init Device Tags

	conerr := d.Reconnect()
	if conerr != nil {
		d.Errorf("First Device connect error: %s", conerr)
		d.DeviceConnected = false
	} else {
		d.DeviceConnected = true
	}

	//Init TagMap

	d.TagMap = make(map[string]string)
	d.TagMap["device"] = d.cfg.Name

	ExtraTags, err := utils.KeyValArrayToMap(d.cfg.ExtraTags)
	if err != nil {
		d.Warnf("Warning on Device  %s Tag gathering: %s", err)
	}
	utils.MapAdd(d.TagMap, ExtraTags)

	// Init stats
	d.InitStats(d.cfg.Name)

	d.SetScanFreq(60)

	return nil
}

// ReleaseClient release connections
func (d *Server) releaseClient() {
	if d.sftpclient != nil {
		d.sftpclient.Close()
	}
	if d.sshclient != nil {
		d.sshclient.Close()
	}
}

// SSHRemoteExec a way to exec basic commands
func (d *Server) SSHRemoteExec(cmd string) (string, error) {
	session, err := d.sshclient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b // get output
	// Finally, run the command TZ

	err = session.Run(cmd)
	if err != nil {
		return b.String(), err
	}
	return b.String(), nil
}

// reconnect does HTTP connection  protocol
func (d *Server) reconnect() error {
	var t time.Duration
	var id string
	var err error
	d.Debugf("Trying Reconnect again....")
	if d.sftpclient != nil {
		d.sftpclient.Close()
	}
	d.sftpclient, d.sshclient, t, id, err = Ping(d.cfg, d.GetLogger(), d.cfg.NmonProtDebug, d.cfg.ID)
	if err != nil {
		d.Errorf("Error on Device connection %s", err)
		return err
	}
	d.Infof("Server Timezone set to: %s", d.Timezone)
	//GetServer TimeZone
	// Create a SSH session. It is one session per command.
	out, err := d.SSHRemoteExec("echo $TZ")
	if err != nil {
		d.Warnf("Error on get TimeZone by cmd echo $TZ: %s", err)
	}
	tz := strings.TrimSuffix(out, "\n")
	if len(tz) > 0 {
		d.Infof("Connected to Device (echo $TZ) OK : ID: %s : Duration %s : Timezone :%s ", id, t.String(), tz)
		_, err := time.LoadLocation(tz)
		if err != nil {
			d.Errorf("Got Timezone: %s is not a valid Timezone : ERR : %s", tz, err)
		} else {
			d.Timezone = tz
			return nil
		}

	}
	out, err = d.SSHRemoteExec("cat /etc/timezone")
	if err != nil {
		d.Warnf("Error on get TimeZone /etc/timezone: %s", err)
		return nil
	}

	tz = strings.TrimSuffix(out, "\n")
	if len(tz) > 0 {
		d.Infof("Connected to Device  (cat /etc/timezone) OK : ID: %s : Duration %s : Timezone :%s ", id, t.String(), tz)
		_, err := time.LoadLocation(tz)
		if err != nil {
			d.Errorf("Got Timezone: %s is not a valid Timezone : ERR : %s", tz, err)
		} else {
			d.Timezone = tz
			return nil
		}
	}
	return nil
}

// checkDeviceConnectivity check if Device connection is ok
func (d *Server) checkDeviceConnectivity() bool {
	d.Debugf("Check Device Connectivity: Nothing to do in the Device %s", d.cfg.ID)

	return true
}
