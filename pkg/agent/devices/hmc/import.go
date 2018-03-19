package hmc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/adejoux/pSeriesCollector/pkg/data/hmcpcm"
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
)

const timeFormat = "2006-01-02T15:04:05-0700"

// GenerateServerMeasurements generate measurements for HMC Managed servers
func (d *HMCServer) GenerateServerMeasurements(pa *pointarray.PointArray, Tags map[string]string, t time.Time, s hmcpcm.ServerData) {

	fieldproc := map[string]interface{}{
		"totalProcUnits":        s.Processor.TotalProcUnits[0],
		"utilizedProcUnits":     s.Processor.UtilizedProcUnits[0],
		"availableProcUnits":    s.Processor.AvailableProcUnits[0],
		"configurableProcUnits": s.Processor.ConfigurableProcUnits[0],
	}

	pa.Append("hmcSystemProcessor", Tags, fieldproc, t)

	fieldmem := map[string]interface{}{
		"totalMem":           s.Memory.TotalMem[0],
		"assignedMemToLpars": s.Memory.AssignedMemToLpars[0],
		"availableMem":       s.Memory.AvailableMem[0],
		"configurableMem":    s.Memory.ConfigurableMem[0],
	}

	pa.Append("hmcSystemMemory", Tags, fieldmem, t)

	for _, spp := range s.SharedProcessorPool {

		SppTags := utils.MapDupAndAdd(Tags, map[string]string{"pool": spp.Name})

		fields := map[string]interface{}{
			"assignedProcUnits":  spp.AssignedProcUnits[0],
			"utilizedProcUnits":  spp.UtilizedProcUnits[0],
			"availableProcUnits": spp.AvailableProcUnits[0],
		}

		pa.Append("hmcSystemSharedProcessorPool", SppTags, fields, t)

	}
}

// GenerateViosMeasurements generate measurementes for VIOS servers
func (d *HMCServer) GenerateViosMeasurements(pa *pointarray.PointArray, Tags map[string]string, t time.Time, v []hmcpcm.ViosData) {

	for _, vios := range v {
		//Check if this vios exist in the device catalog and is enabled
		devcfg, err := db.GetDeviceCfgByID(vios.UUID)
		if err != nil {
			d.Warnf("Any Device in the DB with this name/uuid for VIOS [%s] - [%s]", vios.Name, vios.UUID)
			continue
		}
		if devcfg.EnableHMCStats == false {
			d.Infof("Skeeping Data Importation for Disabled VIOS [%s] - [%s]", vios.Name, vios.UUID)
			continue
		}

		//Getting custom Tags for VIOS

		ViosTags := utils.MapDupAndAdd(Tags, map[string]string{"partition": vios.Name})

		ViosCustomTags, err := utils.KeyValArrayToMap(devcfg.ExtraTags)
		if err != nil {
			d.Warnf("Warning on Device  %s Tag gathering: %s", err)
		}
		utils.MapAdd(ViosTags, ViosCustomTags)

		for _, scsi := range vios.Storage.GenericPhysicalAdapters {

			ScsiTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": scsi.ID})

			fields := map[string]interface{}{
				"transmittedBytes": scsi.TransmittedBytes[0],
				"numOfReads":       scsi.NumOfReads[0],
				"numOfWrites":      scsi.NumOfWrites[0],
				"readBytes":        scsi.ReadBytes[0],
				"writeBytes":       scsi.WriteBytes[0],
			}

			pa.Append("hmcSystemgenericPhysicalAdapters", ScsiTags, fields, t)

		}
		for _, fc := range vios.Storage.FiberChannelAdapters {

			FcTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": fc.ID})

			fields := map[string]interface{}{
				"numOfReads":  fc.NumOfReads[0],
				"numOfWrites": fc.NumOfWrites[0],
				"readBytes":   fc.ReadBytes[0],
				"writeBytes":  fc.WriteBytes[0],
			}

			if len(fc.TransmittedBytes) > 0 {
				fields["transmittedBytes"] = fc.TransmittedBytes[0]
			}
			pa.Append("hmcSystemFiberChannelAdapters", FcTags, fields, t)

		}
		for _, vscsi := range vios.Storage.GenericVirtualAdapters {

			VscsiTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": vscsi.ID})

			fields := map[string]interface{}{
				"numOfReads":  vscsi.NumOfReads[0],
				"numOfWrites": vscsi.NumOfWrites[0],
				"readBytes":   vscsi.ReadBytes[0],
				"writeBytes":  vscsi.WriteBytes[0],
			}

			if len(vscsi.TransmittedBytes) > 0 {
				fields["transmittedBytes"] = vscsi.TransmittedBytes[0]
			}
			pa.Append("hmcSystemGenericVirtualAdapters", VscsiTags, fields, t)

		}
		for _, ssp := range vios.Storage.SharedStoragePools {

			SspTags := utils.MapDupAndAdd(ViosTags, map[string]string{"pool": ssp.ID})

			fields := map[string]interface{}{
				"totalSpace":  ssp.TotalSpace[0],
				"usedSpace":   ssp.UsedSpace[0],
				"numOfReads":  ssp.NumOfReads[0],
				"numOfWrites": ssp.NumOfWrites[0],
				"readBytes":   ssp.ReadBytes[0],
				"writeBytes":  ssp.WriteBytes[0],
			}

			if len(ssp.TransmittedBytes) > 0 {
				fields["transmittedBytes"] = ssp.TransmittedBytes[0]
			}
			pa.Append("hmcSystemSharedStoragePool", SspTags, fields, t)

		}
		for _, net := range vios.Network.GenericAdapters {

			NetTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": net.ID, "type": net.Type})

			fields := map[string]interface{}{
				"receivedPackets": net.ReceivedPackets[0],
				"sentPackets":     net.SentPackets[0],
				"droppedPackets":  net.DroppedPackets[0],
				"sentBytes":       net.SentBytes[0],
				"receivedBytes":   net.ReceivedBytes[0],
			}

			if len(net.TransferredBytes) > 0 {
				fields["transferredBytes"] = net.TransferredBytes[0]
			}
			pa.Append("hmcSystemGenericAdapters", NetTags, fields, t)

		}

		for _, net := range vios.Network.SharedAdapters {

			NetTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": net.ID, "type": net.Type})

			fields := map[string]interface{}{
				"receivedPackets": net.ReceivedPackets[0],
				"sentPackets":     net.SentPackets[0],
				"droppedPackets":  net.DroppedPackets[0],
				"sentBytes":       net.SentBytes[0],
				"receivedBytes":   net.ReceivedBytes[0],
			}

			if len(net.TransferredBytes) > 0 {
				fields["transferredBytes"] = net.TransferredBytes[0]
			}
			pa.Append("hmcSystemSharedAdapters", NetTags, fields, t)
		}
	}
}

// GenerateLparMeasurements generate measurements for LPAR servers
func (d *HMCServer) GenerateLparMeasurements(pa *pointarray.PointArray, Tags map[string]string, t time.Time, l []hmcpcm.LparData) {

	for _, lpar := range l {

		LparTags := utils.MapDupAndAdd(Tags, map[string]string{"partition": lpar.Name})

		fieldproc := map[string]interface{}{
			"maxVirtualProcessors":        lpar.Processor.MaxVirtualProcessors[0],
			"maxProcUnits":                lpar.Processor.MaxProcUnits[0],
			"entitledProcUnits":           lpar.Processor.EntitledProcUnits[0],
			"utilizedProcUnits":           lpar.Processor.UtilizedProcUnits[0],
			"utilizedCappedProcUnits":     lpar.Processor.UtilizedCappedProcUnits[0],
			"utilizedUncappedProcUnits":   lpar.Processor.UtilizedUncappedProcUnits[0],
			"idleProcUnits":               lpar.Processor.IdleProcUnits[0],
			"donatedProcUnits":            lpar.Processor.DonatedProcUnits[0],
			"timeSpentWaitingForDispatch": lpar.Processor.TimeSpentWaitingForDispatch[0],
			"timePerInstructionExecution": lpar.Processor.TimePerInstructionExecution[0],
		}

		pa.Append("hmcPartitionProcessor", LparTags, fieldproc, t)

		fieldmem := map[string]interface{}{
			"logicalMem":        lpar.Memory.LogicalMem[0],
			"backedPhysicalMem": lpar.Memory.BackedPhysicalMem[0],
		}

		pa.Append("hmcPartitionMemory", LparTags, fieldmem, t)

		for _, vfc := range lpar.Storage.VirtualFiberChannelAdapters {

			FcaTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"wwpn":             vfc.Wwpn,
				"physicalPortWWPN": vfc.PhysicalPortWWPN,
				"viosID":           strconv.Itoa(vfc.ViosID),
			})

			fields := map[string]interface{}{
				"transmittedBytes": vfc.TransmittedBytes[0],
				"numOfReads":       vfc.NumOfReads[0],
				"numOfWrites":      vfc.NumOfWrites[0],
				"readBytes":        vfc.ReadBytes[0],
				"writeBytes":       vfc.WriteBytes[0],
			}

			pa.Append("hmcPartitionVirtualFiberChannelAdapters", FcaTags, fields, t)
		}

		for _, vscsi := range lpar.Storage.GenericVirtualAdapters {

			VscsiTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"device": vscsi.ID,
				"viosID": strconv.Itoa(vscsi.ViosID),
			})

			fields := map[string]interface{}{
				"transmittedBytes": vscsi.TransmittedBytes[0],
				"numOfReads":       vscsi.NumOfReads[0],
				"numOfWrites":      vscsi.NumOfWrites[0],
				"readBytes":        vscsi.ReadBytes[0],
				"writeBytes":       vscsi.WriteBytes[0],
			}

			pa.Append("hmcPartitionVSCSIAdapters", VscsiTags, fields, t)

		}

		for _, net := range lpar.Network.VirtualEthernetAdapters {

			NetTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"vlanID":    strconv.Itoa(net.VlanID),
				"vswitchID": strconv.Itoa(net.VswitchID),
				"sea":       net.SharedEthernetAdapterID,
				"viosID":    strconv.Itoa(net.ViosID),
			})

			fields := map[string]interface{}{
				"transferredBytes":         net.TransferredBytes[0],
				"receivedPackets":          net.ReceivedPackets[0],
				"sentPackets":              net.SentPackets[0],
				"droppedPackets":           net.DroppedPackets[0],
				"sentBytes":                net.SentBytes[0],
				"receivedBytes":            net.ReceivedBytes[0],
				"transferredPhysicalBytes": net.TransferredPhysicalBytes[0],
				"receivedPhysicalPackets":  net.ReceivedPhysicalPackets[0],
				"sentPhysicalPackets":      net.SentPhysicalPackets[0],
				"droppedPhysicalPackets":   net.DroppedPhysicalPackets[0],
				"sentPhysicalBytes":        net.SentPhysicalBytes[0],
				"receivedPhysicalBytes":    net.ReceivedPhysicalBytes[0],
			}

			pa.Append("hmcPartitionVirtualEthernetAdapters", NetTags, fields, t)

		}

		for _, net := range lpar.Network.SriovLogicalPorts {

			NetTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"drcIndex":    net.DrcIndex,
				"phyLocation": net.PhysicalLocation,
				"phyDrcIndex": net.PhysicalDrcIndex,
				"phyPortID":   strconv.Itoa(net.PhysicalPortID),
			})

			fields := map[string]interface{}{
				"receivedPackets": net.ReceivedPackets[0],
				"sentPackets":     net.SentPackets[0],
				"droppedPackets":  net.DroppedPackets[0],
				"sentBytes":       net.SentBytes[0],
				"receivedBytes":   net.ReceivedBytes[0],
			}

			pa.Append("hmcPartitionSriovLogicalPorts", NetTags, fields, t)
		}
	}
}

// ScanHMCDevices scan HMC
func (d *HMCServer) ScanHMCDevices() error {
	d.Infof("Scanning  managed systems")

	var err error
	d.System, err = ScanHMC(d.Session)
	if err != nil {
		d.Infof("ERROR on get Managed Systems: %s", err)
		return err
	}
	return nil
}

//ImportData is the entry point for subcommand hmc
func (d *HMCServer) ImportData(points *pointarray.PointArray) error {

	if d.System == nil {
		return fmt.Errorf("Any Scanned SM/LPAR devices detected")
	}

	for _, system := range d.System {

		//Check if this system exist in the device catalog and
		devcfg, err := db.GetDeviceCfgByID(system.UUID)
		if err != nil {
			d.Warnf("Any Device in the DB with this name/uuid for SM [%s] - [%s]", system.SystemName, system.UUID)
			continue
		}
		if devcfg.EnableHMCStats == false {
			d.Infof("Skeeping Data Importation for Disabled SM [%s] - [%s]", system.SystemName, system.UUID)
			continue
		}

		//Getting custom Tags for SM
		Tags := utils.MapDupAndAdd(d.TagMap, map[string]string{"system": system.SystemName})

		SMTags, err := utils.KeyValArrayToMap(devcfg.ExtraTags)
		if err != nil {
			d.Warnf("Warning on Device  %s Tag gathering: %s", err)
		}
		utils.MapAdd(Tags, SMTags)

		//Init SM Data Gathering

		d.Infof("| SYSTEM [%s] | Init data gathering for SM ...", system.SystemName)

		// Get Managed System PCM metrics
		data, dataerr := d.Session.GetSysPCMData(system)
		if dataerr != nil {
			d.Errorf("Error geting PCM data: %s", dataerr)
			continue
		}

		d.Infof("| SYSTEM [%s]  | Processing %d samples ", system.SystemName, len(data.SystemUtil.UtilSamples))

		for _, sample := range data.SystemUtil.UtilSamples {
			timestamp, timeerr := time.Parse(timeFormat, sample.SampleInfo.TimeStamp)
			if timeerr != nil {
				d.Errorf("| SYSTEM [%s] | Error on sample timestamp formating ERROR:%s", system.SystemName, timeerr)
				continue
			}

			switch sample.SampleInfo.Status {
			case 1:
				// if sample sample.SampleInfo.Statusstatus equal 1 we have no data in this sample
				d.Infof(" | SYSTEM [%s] | Skipping sample. Error in sample collection: %s", system.SystemName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
				continue
			case 2:
				// if sample sample.SampleInfo.Statusstatus equal 2 there is some error message but could continue
				d.Warnf(" | SYSTEM [%s] | SAMPLE Status 2: %s", system.SystemName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
			}

			//ServerUtil
			d.GenerateServerMeasurements(points, Tags, timestamp, sample.ServerUtil)

			//ViosUtil
			d.GenerateViosMeasurements(points, Tags, timestamp, sample.ViosUtil)

		}

		if d.ManagedSystemOnly {
			continue
		}

		for _, lpar := range system.Lpars {

			//Check if this system exist in the device catalog and
			devcfg, err := db.GetDeviceCfgByID(lpar.PartitionUUID)
			if err != nil {
				d.Warnf("Any Device in the DB with this name/uuid for LPAR [%s] - [%s]", lpar.PartitionName, lpar.PartitionUUID)
				continue
			}
			if devcfg.EnableHMCStats == false {
				d.Infof("Skeeping Data Importation for Disabled LPAR [%s] - [%s]", lpar.PartitionName, lpar.PartitionUUID)
				continue
			}

			LparTags, err := utils.KeyValArrayToMap(devcfg.ExtraTags)
			if err != nil {
				d.Warnf("Warning on Device  %s Tag gathering: %s", err)
			}
			utils.MapAdd(Tags, LparTags)

			//Init SM Data Gathering

			d.Infof("| SYSTEM [%s] | LPAR [%s] | Init LPAR gathering", system.SystemName, lpar.PartitionName)
			//need to parse the link because the specified hostname can be different
			//of the one specified by the user and the auth cookie will not match

			lparData, lparErr := d.Session.GetLparPCMData(system, lpar)

			if lparErr != nil {
				d.Errorf(" | SYSTEM [%s] | LPAR [%s] | Error geting PCM data: %s", system.SystemName, lpar.PartitionName, lparErr)
				continue
			}

			for _, sample := range lparData.SystemUtil.UtilSamples {

				switch sample.SampleInfo.Status {
				case 1:
					// if sample sample.SampleInfo.Statusstatus equal 1 we have no data in this sample
					d.Infof("| SYSTEM [%s] | LPAR [%s] | Skipping sample. Error in sample collection: %s\n", system.SystemName, lpar.PartitionName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
					continue
				case 2:
					// if sample sample.SampleInfo.Statusstatus equal 2 there is some error message but could continue
					d.Warnf("| SYSTEM [%s] | LPAR [%s] | SAMPLE Status 2: %s", system.SystemName, lpar.PartitionName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
				}

				timestamp, timeerr := time.Parse(timeFormat, sample.SampleInfo.TimeStamp)
				if timeerr != nil {
					d.Errorf("| SYSTEM [%s] | LPAR [%s] | Error on sample timestamp formating ERROR:%s", system.SystemName, lpar.PartitionName, timeerr)
					continue
				}

				//LparUtil
				d.GenerateLparMeasurements(points, Tags, timestamp, sample.LparsUtil)
			}

		}
	}
	return nil
}

//ImportsMData is the entry point for subcommand hmc
func (d *HMCServer) ImportSMData(points *pointarray.PointArray, system *hmcpcm.ManagedSystem) error {

	//Check if this system exist in the device catalog and
	devcfg, err := db.GetDeviceCfgByID(system.UUID)
	if err != nil {
		d.Warnf("Any Device in the DB with this name/uuid for SM [%s] - [%s]", system.SystemName, system.UUID)
		return err
	}
	if devcfg.EnableHMCStats == false {
		d.Infof("Skeeping Data Importation for Disabled SM [%s] - [%s]", system.SystemName, system.UUID)
		return nil
	}

	//Getting custom Tags for SM
	Tags := utils.MapDupAndAdd(d.TagMap, map[string]string{"system": system.SystemName})

	SMTags, err := utils.KeyValArrayToMap(devcfg.ExtraTags)
	if err != nil {
		d.Warnf("Warning on Device  %s Tag gathering: %s", err)
	}
	utils.MapAdd(Tags, SMTags)

	//Init SM Data Gathering

	d.Infof("| SYSTEM [%s] | Init data gathering for SM ...", system.SystemName)

	// Get Managed System PCM metrics
	data, dataerr := d.Session.GetSysPCMData(system)
	if dataerr != nil {
		d.Errorf("Error geting PCM data: %s", dataerr)
		return dataerr
	}

	d.Infof("| SYSTEM [%s]  | Processing %d samples ", system.SystemName, len(data.SystemUtil.UtilSamples))

	for _, sample := range data.SystemUtil.UtilSamples {
		timestamp, timeerr := time.Parse(timeFormat, sample.SampleInfo.TimeStamp)
		if timeerr != nil {
			d.Errorf("| SYSTEM [%s] | Error on sample timestamp formating ERROR:%s", system.SystemName, timeerr)
			continue
		}

		switch sample.SampleInfo.Status {
		case 1:
			// if sample sample.SampleInfo.Statusstatus equal 1 we have no data in this sample
			d.Infof(" | SYSTEM [%s] | Skipping sample. Error in sample collection: %s", system.SystemName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
			continue
		case 2:
			// if sample sample.SampleInfo.Statusstatus equal 2 there is some error message but could continue
			d.Warnf(" | SYSTEM [%s] | SAMPLE Status 2: %s", system.SystemName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
		}

		//ServerUtil
		d.GenerateServerMeasurements(points, Tags, timestamp, sample.ServerUtil)

		//ViosUtil
		d.GenerateViosMeasurements(points, Tags, timestamp, sample.ViosUtil)

	}

	if d.ManagedSystemOnly {
		return nil
	}

	for _, lpar := range system.Lpars {

		//Check if this system exist in the device catalog and
		devcfg, err := db.GetDeviceCfgByID(lpar.PartitionUUID)
		if err != nil {
			d.Warnf("Any Device in the DB with this name/uuid for LPAR [%s] - [%s]", lpar.PartitionName, lpar.PartitionUUID)
			continue
		}
		if devcfg.EnableHMCStats == false {
			d.Infof("Skeeping Data Importation for Disabled LPAR [%s] - [%s]", lpar.PartitionName, lpar.PartitionUUID)
			continue
		}

		LparTags, err := utils.KeyValArrayToMap(devcfg.ExtraTags)
		if err != nil {
			d.Warnf("Warning on Device  %s Tag gathering: %s", err)
		}
		utils.MapAdd(Tags, LparTags)

		//Init SM Data Gathering

		d.Infof("| SYSTEM [%s] | LPAR [%s] | Init LPAR gathering", system.SystemName, lpar.PartitionName)
		//need to parse the link because the specified hostname can be different
		//of the one specified by the user and the auth cookie will not match

		lparData, lparErr := d.Session.GetLparPCMData(system, lpar)

		if lparErr != nil {
			d.Errorf(" | SYSTEM [%s] | LPAR [%s] | Error geting PCM data: %s", system.SystemName, lpar.PartitionName, lparErr)
			continue
		}

		for _, sample := range lparData.SystemUtil.UtilSamples {

			switch sample.SampleInfo.Status {
			case 1:
				// if sample sample.SampleInfo.Statusstatus equal 1 we have no data in this sample
				d.Infof("| SYSTEM [%s] | LPAR [%s] | Skipping sample. Error in sample collection: %s\n", system.SystemName, lpar.PartitionName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
				continue
			case 2:
				// if sample sample.SampleInfo.Statusstatus equal 2 there is some error message but could continue
				d.Warnf("| SYSTEM [%s] | LPAR [%s] | SAMPLE Status 2: %s", system.SystemName, lpar.PartitionName, sample.SampleInfo.ErrorInfo[0].ErrMsg)
			}

			timestamp, timeerr := time.Parse(timeFormat, sample.SampleInfo.TimeStamp)
			if timeerr != nil {
				d.Errorf("| SYSTEM [%s] | LPAR [%s] | Error on sample timestamp formating ERROR:%s", system.SystemName, lpar.PartitionName, timeerr)
				continue
			}

			//LparUtil
			d.GenerateLparMeasurements(points, Tags, timestamp, sample.LparsUtil)
		}

	}

	return nil
}
