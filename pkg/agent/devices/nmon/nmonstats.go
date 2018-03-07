package nmon

import (
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"

	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//PENDING : should we do a more acurated sanitize for field names ???
// "Wait% " => "Wait_percent" ??
// "free(MB)" => "free_mb" ??
// "eth0-read-KB/s" => eth0_read_kb_s ??
// "read/s" => "read_s" ??

func sanitize(in string) string {
	// "User  " => "User"  ??
	return strings.TrimSpace(in)
}

//-----------------------------------------------------------
// Handle CPU Measurements
//
//CPU_ALL,CPU Total XXXX,User%,Sys%,Wait%,Idle%,Busy,PhysicalCPUs
//PCPU_ALL,PCPU Total XXXXX,User  ,Sys  ,Wait  ,Idle  , Entitled Capacity
//SCPU_ALL,SCPU Total XXXXX,User  ,Sys  ,Wait  ,Idle
//CPU0N,CPU N XXXXX,User%,Sys%,Wait%,Idle%
//PCPU0N,PCPU N XXXXX,User ,Sys ,Wait ,Idle
//SCPU01,SCPU N vio4839n2,User ,Sys ,Wait ,Idle

// OLD var cpuallRegexp = regexp.MustCompile(`^CPU\d+|^SCPU\d+|^PCPU\d+`)
var cpuallRegexp = regexp.MustCompile(`^CPU\d+|^SCPU\d+|^PCPU\d+|^+CPU_ALL`)

func (nf *NmonFile) processCPUStats(pa *pointarray.PointArray, Tags map[string]string, t time.Time, lines []string) {

	nf.log.Debugf("Processing CPU stats: %+v", lines)

	for _, line := range lines {
		elems := strings.Split(line, nf.Delimiter)
		name := elems[0]

		tags := utils.MapDupAndAdd(Tags, map[string]string{"cpuid": strings.ToLower(name)})

		fields := make(map[string]interface{})

		for i, value := range elems[2:] {
			if len(nf.DataSeries[name].Columns) < i+1 {
				nf.log.Warnf("Entry added position %d in serie %s since nmon start: skipped COLUMNS [%#+v] Line [%s]", i+1, name, nf.DataSeries[name], line)
				continue
			}

			// try to convert string to integer
			converted, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil || math.IsNaN(converted) {
				nf.log.Warnf("There is some trouble to convert data to float in value [%s] :%s ", value, parseErr)
				//if not working, skip to next value. We don't want text values in InfluxDB.
				continue
			}
			column := nf.DataSeries[name].Columns[i]
			fields[column] = converted

		}

		pa.Append("cpu", tags, fields, t)
	}

}

//-----------------------------------------------------------
// Handle Memory Measurements
// -- On Aix/Vio --
//MEM,Memory XXX,Real Free %,Virtual free %,Real free(MB),Virtual free(MB),Real total(MB),Virtual total(MB)
//MEMNEW,Memory New XXX,Process%,FScache%,System%,Free%,Pinned%,User%
//MEMUSE,Memory Use XXX,%numperm,%minperm,%maxperm,minfree,maxfree,%numclient,%maxclient, lruable pagese
// -- On Linux --
//MEM,Memory MB snmpcoldev01,memtotal,hightotal,lowtotal,swaptotal,memfree,highfree,lowfree,swapfree,memshared,cached,active,bigfree,buffers,swapcached,inactive

//-------------------------------------------------------------------------
var memRegexp = regexp.MustCompile(`^MEM`)

func (nf *NmonFile) processMEMStats(pa *pointarray.PointArray, Tags map[string]string, t time.Time, lines []string) {
	nf.log.Debugf("Processing MEM stats: %+v", lines)

	fields := make(map[string]interface{})

	// no new Tags needed

	for _, line := range lines {
		elems := strings.Split(line, nf.Delimiter)
		name := elems[0]

		for i, value := range elems[2:] {
			if len(nf.DataSeries[name].Columns) < i+1 {
				nf.log.Warnf("Entry added position %d in serie %s since nmon start: skipped COLUMNS [%#+v] LINE [%s] ", i+1, name, nf.DataSeries[name], line)
				continue
			}

			// try to convert string to integer
			converted, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil || math.IsNaN(converted) {
				nf.log.Warnf("There is some trouble to convert data to float in value [%s] :%s ", value, parseErr)
				//if not working, skip to next value. We don't want text values in InfluxDB.
				continue
			}

			// will add prefix on the fields if needed
			column := nf.DataSeries[name].Columns[i]

			var fieldname string
			switch name {

			case "MEMNEW":
				fieldname = "new_" + column
			case "MEMUSE":
				fieldname = "use_" + column
			case "MEM":
				fieldname = column
			default:
				fieldname = column
			}

			fields[fieldname] = converted

		}
	}
	//only one measurement  ( one write) is needed becouse of memnew/memuse/mem has diferent fields and any tag
	pa.Append("memory", Tags, fields, t)
}

// --------------------------------------------
//  Handle TOP Metrics ( Any other not previously handled with Time TAG TXXXX)
//--> NOT ---> TOP,%CPU Utilisation
//TOP,+PID,Time,%CPU,%Usr,%Sys,Size,ResSet,ResText,ResData,ShdLib,MinorFault,MajorFault,Command

var topRegexp = regexp.MustCompile(`^TOP.\d+.(T\d+)`)

func (nf *NmonFile) processTopStats(pa *pointarray.PointArray, Tags map[string]string, t time.Time, lines []string) {
	nf.log.Debugf("Processing TOP stats: %+v", lines)
	for _, line := range lines {
		elems := strings.Split(line, nf.Delimiter)
		name := elems[0]

		if len(elems) < 14 {
			nf.log.Errorf("error TOP import: Elements [%s]", elems)
			return
		}

		fields := make(map[string]interface{})

		var wlmclass string
		if len(elems) < 15 {
			wlmclass = "none"
		} else {
			wlmclass = elems[14]
		}

		tags := utils.MapDupAndAdd(Tags, map[string]string{
			"host":    nf.HostName,
			"pid":     elems[1],
			"command": elems[13],
			"wlm":     wlmclass,
		})

		for i, value := range elems[3:12] {
			column := nf.DataSeries["TOP"].Columns[i]

			if len(nf.Serial) > 0 {
				tags["serial"] = nf.Serial
			}

			// try to convert string to integer
			converted, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil {
				nf.log.Warnf("There is some trouble to convert data to float in value [%s] :%s ", value, parseErr)
				//if not working, skip to next value. We don't want text values in InfluxDB.
				continue
			}

			fields[column] = converted
		}

		pa.Append(name, tags, fields, t)
	}
}

// ------------------------------------------------------------
// Handle Generic Mixed Fields/Tags  in columns for each line
// -----------------------------------------------------------

// Expected:
//--------------------
//- Measurenet NET
//--------------------
var netRegexp = regexp.MustCompile(`^NET([a-zA-Z]*).*`)

// -- AIX/VIO/Linux?---
//NET,Network I/O XXXXX,en10-read-KB/s,lo0-read-KB/s,en10-write-KB/s,lo0-write-KB/s
//NETPACKET,Network Packets XXXXX,en10-reads/s,lo0-reads/s,en10-writes/s,lo0-writes/s
//NETSIZE,Network Size XXXXXX,en10-readsize,lo0-readsize,en10-writesize,lo0-writesize
//NETERROR,Network Errors XXXXXX,en10-ierrs,lo0-ierrs,en10-oerrs,lo0-oerrs,en10-collisions,lo0-collision
// -- Linux ---
//NET,Network I/O XXXXX,lo-read-KB/s,eth0-read-KB/s,lo-write-KB/s,eth0-write-KB/s,
//NETPACKET,Network Packets XXXX,lo-read/s,eth0-read/s,lo-write/s,eth0-write/s,
//--------------------
//- Measurenet SEA
//--------------------
var seaRegexp = regexp.MustCompile(`^SEA([a-zA-Z]*).*`)

//SEA,Shared Ethernet Adapter XXXXXX,ent10-read-KB/s,ent10-write-KB/s,ent11-read-KB/s,ent11-write-KB/s
//SEAPACKET,Shared Ethernet Adapter Packets XXXXXXX,ent10-reads/s,ent10-writes/s,ent11-reads/s,ent11-writes/s
//SEACHPHY,Physical Adapter Traffic Stats XXXXXXXX,ent0_read-KB/s,ent0_write-KB/s,ent0_reads/s,ent0_writes/s,ent0_Transmit_Errors,ent0_Receive_Errors,ent0_Packets_Dropped,ent0_No_ResErrors,ent0_No_Mbuf_Errors,ent2_read-KB/s,ent2_write-KB/s,ent2_reads/s,ent2_writes/s,ent2_Transmit_Errors,ent2_Receive_Errors,ent2_Packets_Dropped,ent2_No_ResErrors,ent2_No_Mbuf_Errors,ent1_read-KB/s,ent1_write-KB/s,ent1_reads/s,ent1_writes/s,ent1_Transmit_Errors,ent1_Receive_Errors,ent1_Packets_Dropped,ent1_No_ResErrors,ent1_No_Mbuf_Errors,ent3_read-KB/s,ent3_write-KB/s,ent3_reads/s,ent3_writes/s,ent3_Transmit_Errors,ent3_Receive_Errors,ent3_Packets_Dropped,ent3_No_ResErrors,ent3_No_Mbuf_Errors
//--------------------
//- Measurenet IOADAP
//--------------------
var ioadaptRegexp = regexp.MustCompile(`^IOADAPT([a-zA-Z]*).*`)

//IOADAPT,Disk Adapter XXXXXX,sissas0_read-KB/s,sissas0_write-KB/s,sissas0_xfer-tps,sissas1_read-KB/s,sissas1_write-KB/s,sissas1_xfer-tps,fcs2_read-KB/s,fcs2_write-KB/s,fcs2_xfer-tps,fcs0_read-KB/s,fcs0_write-KB/s,fcs0_xfer-tps

func (nf *NmonFile) processMixedColumnAsFieldAndTags(pa *pointarray.PointArray, Tags map[string]string, t time.Time, lines []string, measname string, tagname string) {
	nf.log.Debugf("Processing ColumnAsTags  [%s][%s] stats: %+v", measname, tagname, lines)
	//these kind of lines has fields codified in the Line Header, and also
	// example:
	//NETERROR,Network Errors XXXXXX,eth0-ierrs,lo0-ierrs,eth0-oerrs,lo0-oerrs,eth0-collisions,lo0-collision
	// { 2 * { Measurement = net  Fields = "ierrs,oerss,collissions" } and each one with one tag ifname={eth0,lo0}}

	//fist look for tag/field names on all lines

	measurements := make(map[string]map[string]interface{})

	//this regex could generate bugs on systems with docker autogenerated bridgets with name  "br-[NETWORK_ID]"
	var tagfieldRegexp = regexp.MustCompile(`^([^_-]*)[_-]{1}(.*)`)

	for _, line := range lines {
		elems := strings.Split(line, nf.Delimiter)
		name := elems[0]
		for _, col := range nf.DataSeries[name].Columns {
			matched := tagfieldRegexp.FindStringSubmatch(col)
			if len(matched) < 3 {
				nf.log.Warning("There is some trouble on getting tagname-fieldname from column:  %+v, column: %s", matched, col)
				continue
			}
			tag := matched[1]
			field := matched[2]
			if _, ok := measurements[tag]; !ok {
				measurements[tag] = make(map[string]interface{})
			}
			measurements[tag][field] = false
		}
	}
	nf.log.Debugf("Detected Struct %+v", measurements)

	for _, line := range lines {

		elems := strings.Split(line, nf.Delimiter)
		name := elems[0]

		for i, value := range elems[2:] {
			if len(nf.DataSeries[name].Columns) < i+1 {
				nf.log.Warnf("Entry added position %d in serie %s since nmon start: skipped COLUMNS [%#+v] Line [%s]", i+1, name, nf.DataSeries[name], line)
				continue
			}
			column := nf.DataSeries[name].Columns[i]
			matched := tagfieldRegexp.FindStringSubmatch(column)
			if len(matched) < 3 {
				nf.log.Warning("There is some trouble on getting tagname-fieldname from column:  %+v", matched)
				continue
			}
			tag := matched[1]
			field := matched[2]

			// try to convert string to integer
			converted, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil || math.IsNaN(converted) {
				nf.log.Warnf("There is some trouble to convert data to float in value [%s] :%s ", value, parseErr)
				//if not working, skip to next value. We don't want text values in InfluxDB.
				continue
			}
			measurements[tag][field] = converted

		}

	}

	//now we can send all generated data

	for kmeas, meas := range measurements {
		tags := utils.MapDupAndAdd(Tags, map[string]string{tagname: kmeas})
		//TODO clean fields with not float number
		pa.Append(measname, tags, meas, t)
	}
}

//------------------------------------------------------------
// Handle GenericColumnAsTag Lines
// ----------------------------------------------------------
// Expected Measurements
//--------------------
//- Measurenet PAGING
//--------------------
//PAGING,PagingSpace MB Free XXXXXX,hd6,paging00
var pagingRegexp = regexp.MustCompile(`^PAGING([a-zA-Z]*).*`)

//--------------------
//- Measurenet DISK

var diskRegexp = regexp.MustCompile(`^DISK([a-zA-Z]*).*`)

//-- AIX/VIO/Linux--
//DISKBUSYXX,Disk %Busy XXXX,sr0,sda,sda1,sda2,sda5,sdb,dm-0,dm-1
//DISKREADXX,Disk Read KB/s XXXXX,sr0,sda,sda1,sda2,sda5,sdb,dm-0,dm-1
//DISKWRITEXX,Disk Write KB/s XXXXX,sr0,sda,sda1,sda2,sda5,sdb,dm-0,dm-1
//DISKXFERXX,Disk transfers per second XXXXX,hdisk189
//DISKRXFERXX,Transfers from disk (reads) per second XXXXX XXXXX,sr0,sda,sda1,sda2,sda5,sdb,dm-0,dm-1
//DISKBSIZEXX,Disk Block Size XXXXX,sr0,sda,sda1,sda2,sda5,sdb,dm-0,dm-1
//DISKRIOXX,Disk IO Reads per second XXXXX,hdisk189
//DISKWIOXX,Disk IO Writes per second XXXXX,hdisk189
//DISKAVGRIOXX,Disk IO Average Reads per second XXXXX,hdisk0,
//DISKAVGWIOXX,Disk IO Average Writes per second XXXXX,hdisk189
//--------------------
//- Measurenet VG
//--------------------
var vgRegexp = regexp.MustCompile(`^VG([a-zA-Z]*).*`)

// -- AIX/VIO ---
//VGBUSY,Disk Busy Volume Group,rootvg
//VGREAD,Disk Read KB/s Volume Group,rootvg
//VGWRITE,Disk Write KB/s Volume Group,rootvg
//VGXFER,Disk Xfer Volume Group,rootvg
//VGSIZE,Disk Size KB Volume Group,rootvg
//--------------------
//- Measurenet JFS
//--------------------
var jfsRegexp = regexp.MustCompile(`^JFS([a-zA-Z]*).*`)

//JFSFILE,JFS Filespace %Used XXXXXX,/,/home,/usr,/var,/tmp,/admin,/opt,/var/adm/ras/livedump,/logs/system/nmon
//JFSINODE,JFS Inode %Used XXXXX,/,/home,/usr,/var,/tmp,/admin,/opt,/var/adm/ras/livedump,/logs/system/nmon
//--------------------
//- Measurenet FC
//--------------------
var fcRegexp = regexp.MustCompile(`^FC([a-zA-Z]*).*`)

//FCREAD,Fibre Channel Read KB/s,fcs0,fcs2
//FCWRITE,Fibre Channel Write KB/s,fcs0,fcs2
//FCXFERIN,Fibre Channel Tranfers In/s,fcs0,fcs2
//FCXFEROUT,Fibre Channel Tranfers Out/s,fcs0,fcs2
//--------------------
//- Measurenet DG
//--------------------
var dgRegexp = regexp.MustCompile(`^DG([a-zA-Z]*).*`)

//DGBUSY,Disk Group Busy XXXXXXX
//DGREAD,Disk Group Read KB/s XXXXXXX
//DGWRITE,Disk Group Write KB/s XXXXXXXX
//DGSIZE,Disk Group Block Size KB XXXXXXX
//DGXFER,Disk Group Transfers/s XXXXXXX

func (nf *NmonFile) processColumnAsTags(pa *pointarray.PointArray, Tags map[string]string, t time.Time, lines []string, measname string, tagname string, fieldregexp *regexp.Regexp) {
	nf.log.Debugf("Processing ColumnAsTags  [%s][%s] stats: %+v", measname, tagname, lines)
	//these kind of lines has fields codified in the Line Header
	// example: DISKREAD,dddd,a,b,c,d => { 4 * { Measurement = Disk  Field = "read" } and each one with one tag disk={a,b,c,d}}

	//PENDING : to do a buffer to construct only len(nf.DataSeries[name].Columns) measurements

	for _, line := range lines {

		elems := strings.Split(line, nf.Delimiter)
		name := elems[0]

		matched := fieldregexp.FindStringSubmatch(line)
		var fieldname string
		//if not data
		if len(matched[1]) > 0 {
			fieldname = matched[1]
		} else {
			fieldname = matched[0]
		}

		for i, value := range elems[2:] {
			if len(nf.DataSeries[name].Columns) < i+1 {
				nf.log.Warnf("Entry added position %d in serie %s since nmon start: skipped COLUMNS [%#+v] Line [%s]", i+1, name, nf.DataSeries[name], line)
				continue
			}
			column := nf.DataSeries[name].Columns[i]
			columnTag := utils.MapDupAndAdd(Tags, map[string]string{tagname: strings.ToLower(column)})

			// try to convert string to integer
			converted, parseErr := strconv.ParseFloat(value, 64)
			if parseErr != nil || math.IsNaN(converted) {
				nf.log.Warnf("There is some trouble to convert data to float in value [%s] :%s ", value, parseErr)
				//if not working, skip to next value. We don't want text values in InfluxDB.
				continue
			}
			fields := make(map[string]interface{})
			fields[fieldname] = converted
			pa.Append(measname, columnTag, fields, t)
		}

	}
	//only one measurement  ( one write) is needed becouse of memnew/memuse/mem has diferent fields and any tag

}

//-----------------------------------------------------------
// Handle Generic Metrics ( Any other not previously handled with Time TAG TXXXX)
//
//SOMETHING,TXXX , XXXXXXXXX
//
// expected to be handled by these
//--------------------
//- Measurenet POOLS
//--------------------
// -- AIX / VIO--
//POOLS,Multiple CPU Pools XXXX,shcpus_in_sys,max_pool_capacity,entitled_pool_capacity,pool_max_time,pool_busy_time,shcpu_tot_time,shcpu_busy_time,Pool_id,entitled
//--------------------
//- Measurenet LPAR
//--------------------
//LPAR,Logical Partition XXXX,PhysicalCPU,virtualCPUs,logicalCPUs,poolCPUs,entitled,weight,PoolIdle,usedAllCPU%,usedPoolCPU%,SharedCPU,Capped,EC_User%,EC_Sys%,EC_Wait%,EC_Idle%,VP_User%,VP_Sys%,VP_Wait%,VP_Idle%,Folded,Pool_id
//--------------------
//- Measurenet PAGE
//--------------------
//PAGE,Paging XXXX,faults,pgin,pgout,pgsin,pgsout,reclaims,scans,cycles
//--------------------
//- Measurenet PROC
//--------------------
//PROC,Processes XXXX,Runnable,Swap-in,pswitch,syscall,read,write,fork,exec,sem,msg,asleep_bufio,asleep_rawio,asleep_diocio
//--------------------
//- Measurenet AIO
//--------------------
//PROCAIO,Asynchronous I/O XXXX,aioprocs,aiorunning,aiocpu
//--------------------
//- Measurenet FILE
//--------------------
//FILE,File I/O XXXX,iget,namei,dirblk,readch,writech,ttyrawch,ttycanch,ttyoutch
//--------------------
//- Measurenet VM
//--------------------
//VM,Paging and Virtual Memory,nr_dirty,nr_writeback,nr_unstable,nr_page_table_pages,nr_mapped,nr_slab,pgpgin,pgpgout,pswpin,pswpout,pgfree,pgactivate,pgdeactivate,pgfault,pgmajfault,pginodesteal,slabs_scanned,kswapd_steal,kswapd_inodesteal,pageoutrun,allocstall,pgrotated,pgalloc_high,pgalloc_normal,pgalloc_dma,pgrefill_high,pgrefill_normal,pgrefill_dma,pgsteal_high,pgsteal_normal,pgsteal_dma,pgscan_kswapd_high,pgscan_kswapd_normal,pgscan_kswapd_dma,pgscan_direct_high,pgscan_direct_normal,pgscan_direct_dma

var genStatsRegexp = regexp.MustCompile(`\W(T\d{4,16})`)

func (nf *NmonFile) processGenericStats(pa *pointarray.PointArray, Tags map[string]string, t time.Time, line string) {
	nf.log.Debugf("Processing Generic stats: %s", line)

	elems := strings.Split(line, nf.Delimiter)
	name := elems[0]

	//no need to create new tags

	fields := make(map[string]interface{})

	for i, value := range elems[2:] {
		if len(nf.DataSeries[name].Columns) < i+1 {
			nf.log.Warnf("Entry added position %d in serie %s since nmon start: skipped COLUMNS [%#+v] Line [%s]", i+1, name, nf.DataSeries[name], line)
			continue
		}

		// try to convert string to integer
		converted, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil || math.IsNaN(converted) {
			nf.log.Warnf("There is some trouble to convert data to float in value [%s] :%s ", value, parseErr)
			//if not working, skip to next value. We don't want text values in InfluxDB.
			continue
		}
		column := nf.DataSeries[name].Columns[i]
		fields[column] = converted

	}

	measurement := ""
	if nfsRegexp.MatchString(name) || cpuallRegexp.MatchString(name) {
		measurement = name
	} else {
		measurement = nameRegexp.ReplaceAllString(name, "")
	}

	pa.Append(measurement, Tags, fields, t)
}
