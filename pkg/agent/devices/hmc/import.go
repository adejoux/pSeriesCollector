package hmc

import (
	"net/url"
	"strconv"
	"time"

	"github.com/adejoux/pSeriesCollector/pkg/data/hmcpcm"
	"github.com/adejoux/pSeriesCollector/pkg/data/utils"
)

const timeFormat = "2006-01-02T15:04:05-0700"

// GenerateServerMeasurements generate measurements for HMC Managed servers
func (d *HMCServer) GenerateServerMeasurements(pa *PointArray, sysname string, t time.Time, s hmcpcm.ServerData) {

	Tags := utils.MapDupAndAdd(d.TagMap, map[string]string{"system": sysname})

	fieldproc := map[string]interface{}{
		"TotalProcUnits":        s.Processor.TotalProcUnits[0],
		"UtilizedProcUnits":     s.Processor.UtilizedProcUnits[0],
		"availableProcUnits":    s.Processor.AvailableProcUnits[0],
		"configurableProcUnits": s.Processor.ConfigurableProcUnits[0],
	}

	pa.Append("SystemProcessor", Tags, fieldproc, t)

	fieldmem := map[string]interface{}{
		"TotalMem":           s.Memory.TotalMem[0],
		"assignedMemToLpars": s.Memory.AssignedMemToLpars[0],
		"availableMem":       s.Memory.AvailableMem[0],
		"ConfigurableMem":    s.Memory.ConfigurableMem[0],
	}

	pa.Append("SystemMemory", Tags, fieldmem, t)

	for _, spp := range s.SharedProcessorPool {

		SppTags := utils.MapDupAndAdd(Tags, map[string]string{"pool": spp.Name})

		fields := map[string]interface{}{
			"assignedProcUnits":  spp.AssignedProcUnits[0],
			"utilizedProcUnits":  spp.UtilizedProcUnits[0],
			"availableProcUnits": spp.AvailableProcUnits[0],
		}

		pa.Append("SystemSharedProcessorPool", SppTags, fields, t)

	}
}

// GenerateViosMeasurements generate measurementes for VIOS servers
func (d *HMCServer) GenerateViosMeasurements(pa *PointArray, sysname string, t time.Time, v []hmcpcm.ViosData) {

	Tags := utils.MapDupAndAdd(d.TagMap, map[string]string{"system": sysname})

	for _, vios := range v {

		ViosTags := utils.MapDupAndAdd(Tags, map[string]string{"partition": vios.Name})

		for _, scsi := range vios.Storage.GenericPhysicalAdapters {

			ScsiTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": scsi.ID})

			fields := map[string]interface{}{
				"transmittedBytes": scsi.TransmittedBytes[0],
				"numOfReads":       scsi.NumOfReads[0],
				"numOfWrites":      scsi.NumOfWrites[0],
				"readBytes":        scsi.ReadBytes[0],
				"writeBytes":       scsi.WriteBytes[0],
			}

			pa.Append("SystemgenericPhysicalAdapters", ScsiTags, fields, t)

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
			pa.Append("SystemFiberChannelAdapters", FcTags, fields, t)

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
			pa.Append("SystemFiberChannelAdapters", VscsiTags, fields, t)

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
			pa.Append("SystemSharedStoragePool", SspTags, fields, t)

		}
		for _, net := range vios.Network.GenericAdapters {

			NetTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": net.ID, "type": net.Type})

			fields := map[string]interface{}{
				"receivedPackets": net.ReceivedPackets[0],
				"sentPackets":     net.SentPackets[0],
				"droppedPackets":  net.DroppedPackets[0],
				"sentBytes":       net.SentBytes[0],
				"ReceivedBytes":   net.ReceivedBytes[0],
			}

			if len(net.TransferredBytes) > 0 {
				fields["transferredBytes"] = net.TransferredBytes[0]
			}
			pa.Append("SystemGenericAdapters", NetTags, fields, t)

		}

		for _, net := range vios.Network.SharedAdapters {

			NetTags := utils.MapDupAndAdd(ViosTags, map[string]string{"device": net.ID, "type": net.Type})

			fields := map[string]interface{}{
				"receivedPackets": net.ReceivedPackets[0],
				"sentPackets":     net.SentPackets[0],
				"droppedPackets":  net.DroppedPackets[0],
				"sentBytes":       net.SentBytes[0],
				"ReceivedBytes":   net.ReceivedBytes[0],
			}

			if len(net.TransferredBytes) > 0 {
				fields["transferredBytes"] = net.TransferredBytes[0]
			}
			pa.Append("SystemSharedAdapters", NetTags, fields, t)
		}
	}
}

// GenerateLparMeasurements generate measurements for LPAR servers
func (d *HMCServer) GenerateLparMeasurements(pa *PointArray, sysname string, t time.Time, l []hmcpcm.LparData) {

	Tags := utils.MapDupAndAdd(d.TagMap, map[string]string{"system": sysname})

	for _, lpar := range l {

		LparTags := utils.MapDupAndAdd(Tags, map[string]string{"partition": lpar.Name})

		fieldproc := map[string]interface{}{
			"MaxVirtualProcessors":        lpar.Processor.MaxVirtualProcessors[0],
			"MaxProcUnits":                lpar.Processor.MaxProcUnits[0],
			"EntitledProcUnits":           lpar.Processor.EntitledProcUnits[0],
			"UtilizedProcUnits":           lpar.Processor.UtilizedProcUnits[0],
			"UtilizedCappedProcUnits":     lpar.Processor.UtilizedCappedProcUnits[0],
			"UtilizedUncappedProcUnits":   lpar.Processor.UtilizedUncappedProcUnits[0],
			"IdleProcUnits":               lpar.Processor.IdleProcUnits[0],
			"DonatedProcUnits":            lpar.Processor.DonatedProcUnits[0],
			"TimeSpentWaitingForDispatch": lpar.Processor.TimeSpentWaitingForDispatch[0],
			"TimePerInstructionExecution": lpar.Processor.TimePerInstructionExecution[0],
		}

		pa.Append("PartitionProcessor", LparTags, fieldproc, t)

		fieldmem := map[string]interface{}{
			"LogicalMem":        lpar.Memory.LogicalMem[0],
			"BackedPhysicalMem": lpar.Memory.BackedPhysicalMem[0],
		}

		pa.Append("PartitionMemory", LparTags, fieldmem, t)

		for _, vfc := range lpar.Storage.VirtualFiberChannelAdapters {

			FcaTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"wwpn":             vfc.Wwpn,
				"PhysicalPortWWPN": vfc.PhysicalPortWWPN,
				"ViosID":           strconv.Itoa(vfc.ViosID),
			})

			fields := map[string]interface{}{
				"transmittedBytes": vfc.TransmittedBytes[0],
				"numOfReads":       vfc.NumOfReads[0],
				"numOfWrites":      vfc.NumOfWrites[0],
				"readBytes":        vfc.ReadBytes[0],
				"writeBytes":       vfc.WriteBytes[0],
			}

			pa.Append("PartitionVirtualFiberChannelAdapters", FcaTags, fields, t)
		}

		for _, vscsi := range lpar.Storage.GenericVirtualAdapters {

			VscsiTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"device": vscsi.ID,
				"ViosID": strconv.Itoa(vscsi.ViosID),
			})

			fields := map[string]interface{}{
				"transmittedBytes": vscsi.TransmittedBytes[0],
				"numOfReads":       vscsi.NumOfReads[0],
				"numOfWrites":      vscsi.NumOfWrites[0],
				"readBytes":        vscsi.ReadBytes[0],
				"writeBytes":       vscsi.WriteBytes[0],
			}

			pa.Append("PartitionVSCSIAdapters", VscsiTags, fields, t)

		}

		for _, net := range lpar.Network.VirtualEthernetAdapters {

			NetTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"VlanID":    strconv.Itoa(net.VlanID),
				"VswitchID": strconv.Itoa(net.VswitchID),
				"SEA":       net.SharedEthernetAdapterID,
				"ViosID":    strconv.Itoa(net.ViosID),
			})

			fields := map[string]interface{}{
				"transferredBytes":         net.TransferredBytes[0],
				"receivedPackets":          net.ReceivedPackets[0],
				"sentPackets":              net.SentPackets[0],
				"droppedPackets":           net.DroppedPackets[0],
				"sentBytes":                net.SentBytes[0],
				"ReceivedBytes":            net.ReceivedBytes[0],
				"transferredPhysicalBytes": net.TransferredPhysicalBytes[0],
				"receivedPhysicalPackets":  net.ReceivedPhysicalPackets[0],
				"sentPhysicalPackets":      net.SentPhysicalPackets[0],
				"droppedPhysicalPackets":   net.DroppedPhysicalPackets[0],
				"sentPhysicalBytes":        net.SentPhysicalBytes[0],
				"ReceivedPhysicalBytes":    net.ReceivedPhysicalBytes[0],
			}

			pa.Append("PartitionVirtualEthernetAdapters", NetTags, fields, t)

		}

		for _, net := range lpar.Network.SriovLogicalPorts {

			NetTags := utils.MapDupAndAdd(LparTags, map[string]string{
				"DrcIndex":    net.DrcIndex,
				"PhyLocation": net.PhysicalLocation,
				"PhyDrcIndex": net.PhysicalDrcIndex,
				"PhyPortID":   strconv.Itoa(net.PhysicalPortID),
			})

			fields := map[string]interface{}{
				"receivedPackets": net.ReceivedPackets[0],
				"sentPackets":     net.SentPackets[0],
				"droppedPackets":  net.DroppedPackets[0],
				"sentBytes":       net.SentBytes[0],
				"ReceivedBytes":   net.ReceivedBytes[0],
			}

			pa.Append("PartitionSriovLogicalPorts", NetTags, fields, t)
		}
	}
}

//ImportData is the entry point for subcommand hmc
func (d *HMCServer) ImportData(points *PointArray) error {

	d.Infof("Getting list of managed systems")
	systems, err := d.Session.GetManagedSystems()
	if err != nil {
		d.Infof("ERROR on get Managed Systems: %s", err)
		return err
	}
	d.Debugf("ManagedSystems %+v", systems)

	for _, system := range systems {
		//Pending an  easy and powerfull filtering system

		d.Infof("| SYSTEM [%s] | Init processing...", system.Name)
		pcmlinks, syserr := d.Session.GetSystemPCMLinks(system.UUID)
		if syserr != nil {
			d.Infof("Error getting System PCM links: %s", syserr)
			continue
		}
		d.Debugf("| SYSTEM [%s] | Got PCMLinks", system.Name, pcmlinks)

		// Get Managed System PCM metrics
		data, dataerr := d.Session.GetPCMData(pcmlinks.System)
		if dataerr != nil {
			d.Errorf("Error geting PCM data: %s", dataerr)
			continue
		}

		d.Infof("| SYSTEM [%s]  | Processing %d samples ", system.Name, len(data.SystemUtil.UtilSamples))

		for _, sample := range data.SystemUtil.UtilSamples {
			timestamp, timeerr := time.Parse(timeFormat, sample.SampleInfo.TimeStamp)
			if timeerr != nil {
				d.Errorf("| SYSTEM [%s] | Error on sample timestamp formating ERROR:%s", system.Name, timeerr)
				continue
			}

			switch sample.SampleInfo.Status {
			case 1:
				// if sample sample.SampleInfo.Statusstatus equal 1 we have no data in this sample
				d.Infof(" | SYSTEM [%s] | Skipping sample. Error in sample collection: %s", system.Name, sample.SampleInfo.ErrorInfo[0].ErrMsg)
				continue
			case 2:
				// if sample sample.SampleInfo.Statusstatus equal 2 there is some error message but could continue
				d.Warnf(" | SYSTEM [%s] | SAMPLE Status 2: %s", system.Name, sample.SampleInfo.ErrorInfo[0].ErrMsg)
			}

			//ServerUtil
			d.GenerateServerMeasurements(points, system.Name, timestamp, sample.ServerUtil)

			//ViosUtil
			d.GenerateViosMeasurements(points, system.Name, timestamp, sample.ViosUtil)

		}

		if d.ManagedSystemOnly {
			continue
		}

		var lparLinks hmcpcm.PCMLinks
		for _, link := range pcmlinks.Partitions {
			d.Infof("| SYSTEM [%s] | LPAR [%s] | Init LPAR gathering", system.Name, link)
			//need to parse the link because the specified hostname can be different
			//of the one specified by the user and the auth cookie will not match
			rawurl, _ := url.Parse(link)
			var lparGetPCMErr error
			lparLinks, lparGetPCMErr = d.Session.GetPartitionPCMLinks(rawurl.Path)
			if lparGetPCMErr != nil {
				d.Errorf(" | SYSTEM [%s] |LPAR [%s] | Error getting PCM data: %s", system.Name, link, lparGetPCMErr)
				continue
			}
			d.Debugf("| SYSTEM [%s] | LPAR [%s] | Got LPARLinks %+v", system.Name, link, lparLinks)

			for _, lparLink := range lparLinks.Partitions {

				lparData, lparErr := d.Session.GetPCMData(lparLink)

				if lparErr != nil {
					d.Errorf(" | SYSTEM [%s] | LPAR [%s] | Error geting PCM data: %s", system.Name, link, lparErr)
					continue
				}
				d.Infof("")

				for _, sample := range lparData.SystemUtil.UtilSamples {

					switch sample.SampleInfo.Status {
					case 1:
						// if sample sample.SampleInfo.Statusstatus equal 1 we have no data in this sample
						d.Infof("| SYSTEM [%s] | LPAR [%s] | Skipping sample. Error in sample collection: %s\n", system.Name, link, sample.SampleInfo.ErrorInfo[0].ErrMsg)
						continue
					case 2:
						// if sample sample.SampleInfo.Statusstatus equal 2 there is some error message but could continue
						d.Warnf("| SYSTEM [%s] | LPAR [%s] | SAMPLE Status 2: %s", system.Name, link, sample.SampleInfo.ErrorInfo[0].ErrMsg)
					}

					timestamp, timeerr := time.Parse(timeFormat, sample.SampleInfo.TimeStamp)
					if timeerr != nil {
						d.Errorf("| SYSTEM [%s] | LPAR [%s] | Error on sample timestamp formating ERROR:%s", system.Name, link, timeerr)
						continue
					}

					//LparUtil
					d.GenerateLparMeasurements(points, system.Name, timestamp, sample.LparsUtil)
				}
			}
		}
	}
	return nil
}
