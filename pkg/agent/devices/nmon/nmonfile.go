package nmon

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/sftp"

	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
	"github.com/adejoux/pSeriesCollector/pkg/data/rfile"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
)

var hostRegexp = regexp.MustCompile(`^AAA.host.(\S+)`)
var serialRegexp = regexp.MustCompile(`^AAA.SerialNumber.(\S+)`)
var osRegexp = regexp.MustCompile(`^AAA.*(Linux|AIX)`)
var timeRegexp = regexp.MustCompile(`^ZZZZ.(T\d+).(.*)$`)
var intervalRegexp = regexp.MustCompile(`^AAA.interval.(\d+)`)
var headerRegexp = regexp.MustCompile(`^AAA|^BBB|^UARG|\WT\d{4,16}`)
var infoRegexp = regexp.MustCompile(`^AAA.(.*)`)

var delimiterRegexp = regexp.MustCompile(`^\w+(.)`)

// NmonSection data
type NmonSection struct {
	Description string
	Columns     []string
}

// NmonFile type for remote NmonFiles
type NmonFile struct {
	File        *rfile.File
	log         *logrus.Logger
	FilePattern string
	CurFile     string
	Delimiter   string
	//Hostname     string <-not really needed
	//OS           string <-not really needed
	//Serial       string <-not really needed
	TextContent  []string
	Sections     map[string]NmonSection
	sftpConn     *sftp.Client
	HostName     string
	PendingLines []string
	tzLocation   *time.Location
	LastTime     time.Time
	RotateDelay  int
	UserFilters  map[string]*regexp.Regexp
}

// NewNmonFile create a NmonFile , Hostname needed to Parse pattern
func NewNmonFile(sftp *sftp.Client, l *logrus.Logger, pattern string, host string, delay int, filtermap map[string]*regexp.Regexp) *NmonFile {
	return &NmonFile{log: l,
		FilePattern: pattern,
		sftpConn:    sftp,
		HostName:    host,
		RotateDelay: delay,
		UserFilters: filtermap}
}

// AppendText add text section to dashboard
func (nf *NmonFile) AppendText(text string) {
	nf.TextContent = append(nf.TextContent, text)
}

// => check for changes in the file
func (nf *NmonFile) filePathCheck() bool {
	pattern := nf.FilePattern
	t := time.Now()
	nf.log.Debugf("Current Time: %s", t.String())
	then := t.Add(time.Duration(nf.RotateDelay) * time.Second * -1)
	nf.log.Debugf("Time 1 period ago: %s", then.String())
	year, month, day := then.In(nf.tzLocation).Date()
	hour, min, sec := then.In(nf.tzLocation).Clock()
	yearstr := strconv.Itoa(year)
	//  /var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon => (/var/log/nmon/fooserver_20180305_1938.nmon )
	pattern = strings.Replace(pattern, "%{hostname}", strings.ToLower(nf.HostName), -1)
	pattern = strings.Replace(pattern, "%{HOSTNAME}", strings.ToUpper(nf.HostName), -1)
	pattern = strings.Replace(pattern, "%y", yearstr[len(yearstr)-2:], -1) //last two digits
	pattern = strings.Replace(pattern, "%Y", yearstr, -1)
	pattern = strings.Replace(pattern, "%m", fmt.Sprintf("%02d", int(month)), -1)
	pattern = strings.Replace(pattern, "%d", fmt.Sprintf("%02d", day), -1)
	pattern = strings.Replace(pattern, "%H", fmt.Sprintf("%02d", hour), 1)
	pattern = strings.Replace(pattern, "%M", fmt.Sprintf("%02d", min), -1)
	pattern = strings.Replace(pattern, "%S", fmt.Sprintf("%02d", sec), -1)
	if nf.CurFile != pattern {
		nf.log.Infof("Detected Nmon File change OLD [%s] NEW [%s]", nf.CurFile, pattern)
		nf.CurFile = pattern
		return true
	}
	return false
}

// SetPosition  set remote file at newPos Posistion
func (nf *NmonFile) SetPosition(newpos int64) error {
	realpos, err := nf.File.SetPosition(newpos)
	if err != nil {
		nf.log.Debug("Error on set File %s on  expected %d / real %d: error :%s", nf.CurFile, newpos, realpos, err)
		return err
	}
	return nil
}

// Reopen check  and reopen again if needed
func (nf *NmonFile) Reopen() {
	//close remote connection
	nf.File.End()
	//recreate a new connection
	nf.File = rfile.New(nf.sftpConn, nf.log, nf.CurFile)

}

// ReopenIfChanged check if file has changed read last data in previous file and reopen again if needed
func (nf *NmonFile) ReopenIfChanged() bool {

	if nf.filePathCheck() {
		//First we need to read the remaining data in the current buffer
		numlines, filepos := nf.UpdateContent()
		nf.log.Debugf("updated last %d lines ( current filepos in %d)", numlines, filepos)
		//file should be changed (maybe a rotation? or recreation?)
		//close remote connection
		nf.File.End()
		//recreate a new connection
		nf.File = rfile.New(nf.sftpConn, nf.log, nf.CurFile)
		return true
	}
	return false
}

// AddNmonSection add new Section
func (nf *NmonFile) AddNmonSection(line string) bool {
	if len(line) == 0 {
		return false
	}
	if headerRegexp.MatchString(line) {
		nf.log.Debugf("This is line has not a valid Section : Line [%s]", line)
		return false
	}

	/* something happens and is crashing
	badtext := fmt.Sprintf("%s%s", nf.Delimiter, nf.Delimiter)
	badRegexp = regexp.MustCompile(badtext)
	if badRegexp.MatchString(line) {
		continue
	}*/

	elems := strings.Split(line, nf.Delimiter)
	if len(elems) < 3 {
		nf.log.Errorf("ERROR: parsing the following line , not enougth columns (min 3) : %s\n", line)
		return false
	}
	name := elems[0]
	desc := elems[1]

	nf.log.Debugf("Adding Section [%s] - %s\n", name, desc)
	sec := nf.Sections[name]
	sec.Description = desc
	for _, v := range elems[2:] {
		sec.Columns = append(sec.Columns, sanitize(v))
	}
	nf.Sections[name] = sec
	return true
}

//PENDING : should we do a more acurated sanitize for field names ???
// "Wait% " => "Wait_percent" ??
// "free(MB)" => "free_mb" ??
// "eth0-read-KB/s" => eth0_read_kb_s ??
// "read/s" => "read_s" ??

func sanitize(in string) string {
	// "User  " => "User"  ??
	return strings.TrimSpace(in)
}

const maxInitSectionTimes = 20

// InitSectionDefs Initialize section definitions.
func (nf *NmonFile) InitSectionDefs() (int64, error) {
	//Map init
	nf.Sections = make(map[string]NmonSection)
	// Get Content as many  times as needed until a ZZZZ section found
	loopcontrol := 0
	for {
		data, pos, err := nf.File.ContentUntilMatch(timeRegexp)
		if err != nil {
			return 0, err
		}
		nf.log.Infof("InitSectionDefs: Initialice NMONFILE:  %s | Section Header length: (%d) ", nf.CurFile, len(data))

		first := true

		for k, line := range data {
			//var badRegexp *regexp.Regexp
			//Look for Nmon Delimiter

			if first {
				if delimiterRegexp.MatchString(line) {
					matched := delimiterRegexp.FindStringSubmatch(line)
					nf.Delimiter = matched[1]
				} else {
					nf.Delimiter = ","
				}
				nf.log.Debugf("InitSectionDefs: NMONFILE DELIMITER SET: [%s]", nf.Delimiter)

				first = false
			}
			// replace data if needed depending on the delimiter
			if nf.Delimiter == ";" {
				line = strings.Replace(line, ",", ".", -1)
				data[k] = line
			}
			//begin data check
			nf.log.Debugf("InitSectionDefs:NMONFILE(%d): %s", k, line)

			//if time Line reached we will finish our header check but other headers could appear
			if timeRegexp.MatchString(line) {
				matched := timeRegexp.FindStringSubmatch(line)
				nf.log.Debugf("InitSectionDefs: Found Time ID [%s] :  %s", matched[1], matched[2])
				//nf.TimeStamps[matched[1]] = matched[2]
				// from now all lines will have data to

				nf.PendingLines = append(nf.PendingLines, data[k:]...)
				return pos, nil
				//break
			}

			/* while not really needed we will disable these data
			if hostRegexp.MatchString(line) {
				matched := hostRegexp.FindStringSubmatch(line)
				nf.Hostname = strings.ToLower(matched[1])
				continue
			}

			if serialRegexp.MatchString(line) {
				matched := serialRegexp.FindStringSubmatch(line)
				nf.Serial = strings.ToLower(matched[1])
				continue
			}

			if osRegexp.MatchString(line) {
				matched := osRegexp.FindStringSubmatch(line)
				nf.OS = strings.ToLower(matched[1])
				continue
			}*/

			if infoRegexp.MatchString(line) {
				matched := infoRegexp.FindStringSubmatch(line)
				nf.AppendText(matched[1])
				continue
			}
			nf.AddNmonSection(line)

		}

		nf.log.Warnf("Any ZZZZ section found on (%d) lines of the file ", len(data))
		// this sleep prevents for and infinite loop while waiting for nmon process
		// finish the header part of the nmon file.
		// is a workarround until I can think of something better
		loopcontrol++
		if loopcontrol > maxInitSectionTimes {

			return 0, fmt.Errorf("Too many times trying to read the file and  looking for an ZZZZ section, aborting.... ")
		}
		time.Sleep(time.Duration(5 * time.Second))
	}

}

// Init Initialize NmonFile struct return current position after initialized
func (nf *NmonFile) Init(timezone string) (int64, error) {
	nf.SetTimeZoneLocation(timezone) //pending set location from system
	nf.filePathCheck()
	nf.File = rfile.New(nf.sftpConn, nf.log, nf.CurFile)
	pos, err := nf.InitSectionDefs()
	nf.log.Debugf("Init: End of NMONFile %s Initialization,  pending lines on buffer: [%d] Current file position: [%d] pending [%+v]", nf.FilePattern, len(nf.PendingLines), pos, nf.PendingLines)
	return pos, err
}

// UpdateContent from remoteFile return num of  new lines , and new pos
func (nf *NmonFile) UpdateContent() (int, int64) {
	morelines, pos := nf.File.Content()
	nf.log.Infof("UpdateContent: Got new %d lines from NmonFile ", len(morelines))
	// replace data if needed depending on the delimiter
	if nf.Delimiter == ";" {
		for k, line := range morelines {
			line = strings.Replace(line, ",", ".", -1)
			morelines[k] = line
		}
	}
	nf.PendingLines = append(nf.PendingLines, morelines...)
	return len(morelines), pos
}

const timeformat = "15:04:05 02-Jan-2006"

//SetTimeZoneLocation set the timezone used to input metrics in InfluxDB
func (nf *NmonFile) SetTimeZoneLocation(tz string) (err error) {

	var loc *time.Location
	if len(tz) > 0 {
		loc, err = time.LoadLocation(tz)
		nf.log.Infof("END Timezone Finaly set at %s", tz)
		if err != nil {
			loc = time.FixedZone("Europe/Paris", 2*60*60)
			nf.log.Warnf("WARN On set timezone %s, %s", tz, err)
		}
	} else {
		timezone, _ := time.Now().In(time.Local).Zone()
		nf.log.Infof("END Timezone Finaly set at %s", timezone)
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			loc = time.FixedZone("Europe/Paris", 2*60*60)
			nf.log.Warnf("WARN On set timezone %s, %s", timezone, err)
		}
	}

	nf.tzLocation = loc
	return
}

func (nf *NmonFile) convertTimeStamp(s string) (time.Time, error) {
	var err error
	if s == "now" {
		return time.Now().Truncate(24 * time.Hour), err
	}

	//replace separator
	if len(s) < 9 {
		return time.Now(), fmt.Errorf("SetTimeZoneLocation: too small timestamp string to convert : %s", s)
	}
	stamp := s[0:8] + " " + s[9:]
	t, err := time.ParseInLocation(timeformat, stamp, nf.tzLocation)
	return t, err
}

//ResetPending remove buffered data
func (nf *NmonFile) ResetPending() {
	nf.log.Debugf("ResetPending:Reseting current Buffer containing (%d) lines [%+v]", len(nf.PendingLines), nf.PendingLines)
	nf.PendingLines = []string{}
}

//ResetText remove buffered data
func (nf *NmonFile) ResetText() {
	nf.log.Debugf("ResetText:Reseting current Text Buffer containing (%d) ", len(nf.TextContent))
	nf.TextContent = []string{}
}

// ProcessPending process last
func (nf *NmonFile) ProcessPending(points *pointarray.PointArray, tags map[string]string) {
	var tsID string
	var ts string
	nf.log.Debug("ProcessPending: Init")

	if len(nf.PendingLines) == 0 {
		nf.log.Warnf("nothing to process in the process pending buffer")
		return
	}

	//do while  no more ZZZ section  found
	for {
		// first line should be a ZZZZ section
		firstline := nf.PendingLines[0]
		if timeRegexp.MatchString(firstline) {
			matched := timeRegexp.FindStringSubmatch(firstline)
			tsID = matched[1]
			ts = matched[2]
			nf.log.Debugf("ProcessPending: Found Time ID [%s] :  %s", matched[1], matched[2])

		} else {
			nf.log.Errorf("ProcessPending: ERROR: first Pending data is not ZZZZZ (Time) section got this one [%s]", firstline)
			if len(nf.PendingLines) > 1 {
				nf.PendingLines = nf.PendingLines[1:]
			}
			//PENDING what to do if this happens?
			return
		}

		//a complete set of data
		var nmonChunk []string
		last := 0
		for i := 1; i < len(nf.PendingLines); i++ {
			line := nf.PendingLines[i]
			if timeRegexp.MatchString(line) {
				//if another ZZZZZ end process
				last = i
				break
			}
			nf.log.Debugf("ProcessPending: Line (%d) : %s", i, line)
			//no XXXXX,TimeID,
			nmonChunk = append(nmonChunk, line)
		}
		//rewrite pending lines
		nf.PendingLines = nf.PendingLines[last:]
		t, err := nf.convertTimeStamp(ts)
		if err != nil {
			nf.log.Errorf("ProcessPending: Error on Timestamp conversion %s", err)
			return
		}
		nf.ProcessChunk(points, tags, t, tsID, nmonChunk)
		nf.LastTime = t
		if last == 0 {
			nf.log.Debugf("ProcessPending: no more lines in pending lines buffer. Exiting...")
			//no more ZZZ in the remaining Lines
			return
		}
	}

}

//Original from nmon2influxdb was regexp.MustCompile(`T0+\W|^Z|^TOP.%CPU`)
//not really filtering TOP on some devices
//we have added a UserFilter per device
var skipRegexp = regexp.MustCompile(`T0+\W|^Z|^TOP.%CPU`)

// ProcessChunk process a
func (nf *NmonFile) ProcessChunk(pa *pointarray.PointArray, Tags map[string]string, t time.Time, timeID string, lines []string) {
	nf.log.Infof("ProcessChunk: Decoding Chunk for Timestamp %s  with %d Elements ", t.String(), len(lines))

	regstr := ""
	for _, line := range lines {
		//check if exit header to process data
		header := strings.Split(line, nf.Delimiter)[0]
		if _, ok := nf.Sections[header]; !ok {
			nf.log.Infof("ProcessChunk: Line  not in Header [%s] trying to add...", line)
			// if not perhaps is a new header
			if nf.AddNmonSection(line) == true {
				if len(regstr) > 0 {
					regstr = regstr + "|^" + header
				} else {
					regstr = "^" + header
				}
			}
			continue
		}
	}
	if len(regstr) > 0 {
		//there is a new
		nf.log.Infof("ProcessChunk: Found not allowed sections REGEX = [%s]", regstr)
		contains, notcontains := utils.Grep(lines, regexp.MustCompile(regstr))
		lines = notcontains
		nf.log.Debugf("ProcessChunk: CONTAINS:%+v", contains)
		nf.log.Debugf("ProcessChunk: NOTCONTAINS: %+v", notcontains)
	}

	remain := lines
	var linesok []string
	var linesnotok []string

	for {
		if len(remain) == 0 {
			//exit from the loop if any other line pending to  process.
			break
		}
		//Filter HardCoded SkipSections
		_, remain = utils.Grep(remain, skipRegexp)
		//Filter HardCoded HeaderSections
		remain, _ = utils.Grep(remain, headerRegexp)
		//Filter Not In Time data
		remain, linesnotok = utils.Grep(remain, regexp.MustCompile(`\W`+timeID))
		if len(linesnotok) > 0 {
			nf.log.Warnf("ProcessChunk: Lines not in time  TIMEID [%s] : Lines :[%+v]", timeID, linesnotok)
		}
		//-----------------------------------------------------------------------------
		// We will only process , format and send measurements from known Nmon Seccions
		//-----------------------------------------------------------------------------
		//USER FILTERS

		if len(nf.UserFilters) > 0 {
			nf.log.Infof("Found %d user Filters", len(nf.UserFilters))
			var userfiltered []string
			for k, v := range nf.UserFilters {
				userfiltered, remain = utils.Grep(remain, v)
				nf.log.Infof("USER Filter %s APPLIED: Skipped %d Lines", k, len(userfiltered))
				nf.log.Debugf("USER Filter %s APPLIED: Skipped: %+v", k, userfiltered)
			}
		}

		//CPU
		linesok, remain = utils.Grep(remain, cpuRegexp)
		if len(linesok) > 0 {
			nf.processCPUStats(pa, Tags, t, linesok)
		}
		if len(remain) == 0 {
			break
		}
		// MEM Stats
		linesok, remain = utils.Grep(remain, memRegexp)
		if len(linesok) > 0 {
			nf.processMEMStats(pa, Tags, t, linesok)
		}
		if len(remain) == 0 {
			break
		}
		// PAGING Stats
		linesok, remain = utils.Grep(remain, pagingRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsTags(pa, Tags, t, linesok, "nmonPaging", "psname", pagingRegexp)
		}
		if len(remain) == 0 {
			break
		}
		// DISK Stats
		linesok, remain = utils.Grep(remain, diskRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsTags(pa, Tags, t, linesok, "nmonDisks", "diskname", diskRegexp)
		}
		if len(remain) == 0 {
			break
		}
		// VG Stats
		linesok, remain = utils.Grep(remain, vgRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsTags(pa, Tags, t, linesok, "nmonVolumegroup", "vgname", vgRegexp)
		}
		if len(remain) == 0 {
			break
		}
		// JFS Stats
		linesok, remain = utils.Grep(remain, jfsRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsTags(pa, Tags, t, linesok, "nmonJfs", "fsname", jfsRegexp)
		}
		if len(remain) == 0 {
			break
		}
		// FC Stats
		linesok, remain = utils.Grep(remain, fcRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsTags(pa, Tags, t, linesok, "nmonFiberchannel", "fcname", fcRegexp)
		}
		if len(remain) == 0 {
			break
		}
		// DG Stats
		linesok, remain = utils.Grep(remain, dgRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsTags(pa, Tags, t, linesok, "nmonDiskgroup", "dgname", dgRegexp)
		}
		if len(remain) == 0 {
			break
		}
		//NET stats
		linesok, remain = utils.Grep(remain, netRegexp)
		if len(linesok) > 0 {
			nf.processMixedColumnAsFieldAndTags(pa, Tags, t, linesok, "nmonNetwork", "ifname", regexp.MustCompile(`^(br-[^_-]*|[^_-]*)[_-]{1}(.*)`))
		}
		if len(remain) == 0 {
			break
		}
		//SEA stats
		linesok, remain = utils.Grep(remain, seaRegexp)
		if len(linesok) > 0 {
			nf.processMixedColumnAsFieldAndTags(pa, Tags, t, linesok, "nmonSea", "seaname", nil)
		}
		if len(remain) == 0 {
			break
		}
		//IOADAPT stats
		linesok, remain = utils.Grep(remain, ioadaptRegexp)
		if len(linesok) > 0 {
			nf.processMixedColumnAsFieldAndTags(pa, Tags, t, linesok, "nmonIoadapt", "adaptname", nil)
		}
		if len(remain) == 0 {
			break
		}
		//TOP stats
		linesok, remain = utils.Grep(remain, topRegexp)
		if len(linesok) > 0 {
			nf.processTopStats(pa, Tags, t, linesok)
		}
		if len(remain) == 0 {
			break
		}
		//POOLS,LPAR,PAGE,PROC,PROCAIO,FILE,VM
		linesok, remain = utils.Grep(remain, columAsFieldRegexp)
		if len(linesok) > 0 {
			nf.processColumnAsField(pa, Tags, t, linesok)
		}
		if len(remain) != 0 {
			nf.log.Warnf("ProcessChunk: Lines not processed [%+v] Perhaps is not in Catalog????...", remain)
			break
		}
	}

}
