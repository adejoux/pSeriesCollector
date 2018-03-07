package nmon

import (
	"github.com/Sirupsen/logrus"

	"github.com/adejoux/pSeriesCollector/pkg/data/rfile"

	"github.com/pkg/sftp"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
)

var hostRegexp = regexp.MustCompile(`^AAA.host.(\S+)`)
var serialRegexp = regexp.MustCompile(`^AAA.SerialNumber.(\S+)`)
var osRegexp = regexp.MustCompile(`^AAA.*(Linux|AIX)`)
var timeRegexp = regexp.MustCompile(`^ZZZZ.(T\d+).(.*)$`)
var intervalRegexp = regexp.MustCompile(`^AAA.interval.(\d+)`)
var headerRegexp = regexp.MustCompile(`^AAA|^BBB|^UARG|\WT\d{4,16}`)
var infoRegexp = regexp.MustCompile(`^AAA.(.*)`)

var skipRegexp = regexp.MustCompile(`T0+\W|^Z|^TOP.%CPU`)

var nfsRegexp = regexp.MustCompile(`^NFS`)
var nameRegexp = regexp.MustCompile(`(\d+)$`)

var delimiterRegexp = regexp.MustCompile(`^\w+(.)`)

// DataSerie data
type DataSerie struct {
	Columns []string
}

// NmonFile type for remote NmonFiles
type NmonFile struct {
	File         *rfile.File
	log          *logrus.Logger
	FilePattern  string
	CurFile      string
	Delimiter    string
	Hostname     string
	OS           string
	Serial       string
	TextContent  string
	DataSeries   map[string]DataSerie
	sftpConn     *sftp.Client
	HostName     string
	PendingLines []string
	tzLocation   *time.Location
}

// NewNmonFile create a NmonFile
func NewNmonFile(sftp *sftp.Client, l *logrus.Logger, pattern string, host string) *NmonFile {
	return &NmonFile{log: l, FilePattern: pattern, sftpConn: sftp, HostName: host}
}

// AppendText add text section to dashboard
func (nf *NmonFile) AppendText(text string) {
	nf.TextContent += text
}

// => l
func (nf *NmonFile) filePathCheck() bool {
	pattern := nf.FilePattern
	t := time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	//  /var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon => (/var/log/nmon/fooserver_20180305_1938.nmon )
	pattern = strings.Replace(pattern, "%{hostname}", strings.ToLower(nf.HostName), -1)
	pattern = strings.Replace(pattern, "%{HOSTNAME}", strings.ToUpper(nf.HostName), -1)
	pattern = strings.Replace(pattern, "%Y", strconv.Itoa(year), -1)
	pattern = strings.Replace(pattern, "%m", strconv.Itoa(int(month)), -1)
	pattern = strings.Replace(pattern, "%d", strconv.Itoa(day), -1)
	pattern = strings.Replace(pattern, "%H", strconv.Itoa(hour), 1)
	pattern = strings.Replace(pattern, "%M", strconv.Itoa(min), -1)
	pattern = strings.Replace(pattern, "%S", strconv.Itoa(sec), -1)
	if nf.CurFile != pattern {
		nf.log.Debugf("Detected Nmon File change OLD [%s] NEW [%s]", nf.CurFile, pattern)
		nf.CurFile = pattern
		return true
	}
	return false
}

// CheckFile check if file has changed and reopen again if needed
func (nf *NmonFile) CheckFile() {
	if nf.filePathCheck() {
		//file should be changed (maybe a rotation? or recreation?)
		//close remote connection
		nf.File.End()
		//recreate a new connection
		nf.File = rfile.New(nf.sftpConn, nf.log, nf.CurFile)
		//initializing file
	}
}

// AddNmonSection add new Section
func (nf *NmonFile) AddNmonSection(line string) {
	if len(line) == 0 {
		return
	}
	if headerRegexp.MatchString(line) {
		nf.log.Debug("This is not a valid Header Line [%d]")
		return
	}

	/* something happens and is crashing
	badtext := fmt.Sprintf("%s%s", nf.Delimiter, nf.Delimiter)
	badRegexp = regexp.MustCompile(badtext)
	if badRegexp.MatchString(line) {
		continue
	}*/

	elems := strings.Split(line, nf.Delimiter)
	if len(elems) < 3 {
		nf.log.Errorf("ERROR: parsing the following line : %s\n", line)
		return
	}
	name := elems[0]

	nf.log.Debugf("Adding serie %s\n", name)
	dataserie := nf.DataSeries[name]
	//dataserie.Columns = elems[2:]
	for _, v := range elems[2:] {
		dataserie.Columns = append(dataserie.Columns, sanitize(v))
	}
	nf.DataSeries[name] = dataserie
}

// Init Initialize NmonFile struct
func (nf *NmonFile) Init() {
	nf.SetLocation("") //pending set location from system
	nf.filePathCheck()
	nf.File = rfile.New(nf.sftpConn, nf.log, nf.CurFile)
	//Map init
	nf.DataSeries = make(map[string]DataSerie)
	// Get Content
	data := nf.File.Content()
	nf.log.Infof("Initialice NMONFILE: %s", nf.FilePattern)

	first := true

	// PENDIND : add a way to skip some metrics
	last := 0
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
			nf.log.Debugf("NMONFILE DELIMITER SET: [%s]", nf.Delimiter)

			first = false
		}
		// replace data if needed depending on the delimiter
		if nf.Delimiter == ";" {
			line = strings.Replace(line, ",", ".", -1)
			data[k] = line
		}
		//begin data check
		nf.log.Debugf("NMONFILE(%d): %s", k, line)

		//if time Line reached we will finish our header check but other headers could appear
		if timeRegexp.MatchString(line) {
			matched := timeRegexp.FindStringSubmatch(line)
			nf.log.Debugf("Found Time ID [%s] :  %s", matched[1], matched[2])
			//nf.TimeStamps[matched[1]] = matched[2]
			// from now all lines will have data to
			last = k
			break
		}

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
		}

		if infoRegexp.MatchString(line) {
			matched := infoRegexp.FindStringSubmatch(line)
			nf.AppendText(matched[1])
			continue
		}
		nf.AddNmonSection(line)

	}
	// a time  ZZZZ section has been reached on line

	nf.PendingLines = append(nf.PendingLines, data[last:]...)

	nf.log.Debugf("End of NMONFile %s Initialization,  pending lines %d", nf.FilePattern, len(nf.PendingLines))
}

// UpdateContent from remoteFile
func (nf *NmonFile) UpdateContent() {
	morelines := nf.File.Content()
	nf.log.Infof("Got new %d lines from NmonFile ", len(morelines))
	// replace data if needed depending on the delimiter
	if nf.Delimiter == ";" {
		for k, line := range morelines {
			line = strings.Replace(line, ",", ".", -1)
			morelines[k] = line
		}
	}
	nf.PendingLines = append(nf.PendingLines, morelines...)
}

const timeformat = "15:04:05 02-Jan-2006"

//SetLocation set the timezone used to input metrics in InfluxDB
func (nf *NmonFile) SetLocation(tz string) (err error) {
	var loc *time.Location
	if len(tz) > 0 {
		loc, err = time.LoadLocation(tz)
		if err != nil {
			loc = time.FixedZone("Europe/Paris", 2*60*60)
		}
	} else {
		timezone, _ := time.Now().In(time.Local).Zone()
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			loc = time.FixedZone("Europe/Paris", 2*60*60)
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
	stamp := s[0:8] + " " + s[9:]
	t, err := time.ParseInLocation(timeformat, stamp, nf.tzLocation)
	return t, err
}

// ProcessPending process last
func (nf *NmonFile) ProcessPending(points *pointarray.PointArray, tags map[string]string) {
	var tsID string
	var ts string
	nf.log.Debug("Processing Pending Lines")

	//do while  no more ZZZ section  found
	for {
		// first line should be a ZZZZ section
		firstline := nf.PendingLines[0]
		if timeRegexp.MatchString(firstline) {
			matched := timeRegexp.FindStringSubmatch(firstline)
			tsID = matched[1]
			ts = matched[2]
			nf.log.Debugf("Found Time ID [%s] :  %s", matched[1], matched[2])

		} else {
			nf.log.Errorf("ERROR: first Pending data is not ZZZZZ (Time) section")
		}

		//a complete set of data
		var nmonChunk []string
		last := 0
		for i := 1; i < len(nf.PendingLines); i++ {
			line := nf.PendingLines[i]
			if timeRegexp.MatchString(line) {
				last = i
				break
			}
			nf.log.Debugf("Line (%d) : %s", i, line)
			//if another ZZZZZ end process
			//no XXXXX,TimeID,
			nmonChunk = append(nmonChunk, line)
		}
		nf.PendingLines = nf.PendingLines[last:]
		if last == 0 {
			//no more ZZZ in the remaining Lines
			return
		}
		t, err := nf.convertTimeStamp(ts)
		if err != nil {
			nf.log.Errorf("Error on Timestamp conversion %s", err)
			continue
		}
		nf.ProcessChunk(points, tags, t, tsID, nmonChunk)
	}

}

// ProcessChunk process a
func (nf *NmonFile) ProcessChunk(pa *pointarray.PointArray, Tags map[string]string, t time.Time, timeID string, lines []string) {
	nf.log.Infof("Decoding Chunk for Timestamp %s  with %d Elements ", t.String(), len(lines))

	for _, line := range lines {
		//check if exit header to process data
		header := strings.Split(line, nf.Delimiter)[0]
		if _, ok := nf.DataSeries[header]; !ok {
			nf.log.Infof("Line  not in Header [%s] tring to add...", line)
			// if not perhaps is a new header
			nf.AddNmonSection(line)
			continue
		}
		//PENDING a way to skip metrics
		if skipRegexp.MatchString(line) {
			continue
		}
		// CPU Stats
		if cpuallRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processCPUStats(pa, Tags, t, []string{line})
			continue
		}
		// MEM Stats
		if memRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processMEMStats(pa, Tags, t, []string{line})
			continue
		}
		// PAGING Stats
		if pagingRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processColumnAsTags(pa, Tags, t, []string{line}, "paging", "psname", pagingRegexp)
			continue
		}
		// DISK Stats
		if diskRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processColumnAsTags(pa, Tags, t, []string{line}, "disks", "diskname", diskRegexp)
			continue
		}
		// VG Stats
		if vgRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processColumnAsTags(pa, Tags, t, []string{line}, "volumegroup", "vgname", vgRegexp)
			continue
		}
		// JFS Stats
		if jfsRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processColumnAsTags(pa, Tags, t, []string{line}, "jfs", "fsname", jfsRegexp)
			continue
		}
		// FC Stats
		if fcRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processColumnAsTags(pa, Tags, t, []string{line}, "fiberchannel", "fcname", fcRegexp)
			continue
		}
		// DG Stats
		if dgRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processColumnAsTags(pa, Tags, t, []string{line}, "diskgroup", "dgname", dgRegexp)
			continue
		}

		//NET stats
		if netRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processMixedColumnAsFieldAndTags(pa, Tags, t, []string{line}, "network", "ifname")
			continue
		}

		//SEA stats
		if seaRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processMixedColumnAsFieldAndTags(pa, Tags, t, []string{line}, "sea", "seaname")
			continue
		}

		//IOADAPT stats
		if ioadaptRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processMixedColumnAsFieldAndTags(pa, Tags, t, []string{line}, "ioadapt", "adaptname")
			continue
		}

		//TOP stats
		if topRegexp.MatchString(line) {
			matched := topRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processTopStats(pa, Tags, t, []string{line})
			continue
		}

		//Other Generic Stats
		if genStatsRegexp.MatchString(line) {
			matched := genStatsRegexp.FindStringSubmatch(line)
			if matched[1] != timeID {
				nf.log.Warning("Line not in time  TIMEID [%s] : Line :[%s]", timeID, line)
				continue
			}
			nf.processGenericStats(pa, Tags, t, line)
			continue
		}

		nf.log.Warnf("Line not processed [%s] adding to the header definitions...", line)

	}
}
