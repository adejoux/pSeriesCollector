package nmon

import (
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
)

// ImportData getmon N data from Remote devices
func (d *Server) ImportData(points *pointarray.PointArray) error {

	d.Infof("ImportData: Import Nmon data on remote device (%s) ", d.cfg.NmonIP)
	if d.NmonFile == nil {
		d.Infof("ImportData: Initializing Nmon Remote File")
		nf := NewNmonFile(d.sftpclient, d.GetLogger(), d.cfg.NmonFilePath, d.cfg.Name, d.cfg.NmonFreq, d.FilterMap)
		filepos, err := nf.Init(d.Timezone)
		if err != nil {
			d.Errorf("ImportData: Something happen on Initialize Nmon file: %s [Reopen Again....]", err)
			return err
		}
		d.NmonFile = nf
		// Got last known position
		info, err := db.GetNmonFileInfoByIDFile(d.cfg.ID, d.NmonFile.CurFile)
		if err != nil {
			d.Debugf("ImportData: Warning on get file info for ID [%s] and file [%s] ", d.cfg.ID, d.NmonFile.CurFile)
			d.Infof("ImportData: Current File Position %s is: %d", d.NmonFile.CurFile, filepos)
		} else {
			//	d.NmonFile.ResetPending()
			d.NmonFile.SetPosition(info.LastPosition)
			d.Infof("ImportData: Updated File Position %s now to: %d", d.NmonFile.CurFile, info.LastPosition)
		}
		d.Debugf("ImportData: Found Sections: %#+v", d.NmonFile.Sections)
		d.Debugf("ImportData: Found Content %s", d.NmonFile.TextContent)
	}

	if d.NmonFile.ReopenIfChanged() {
		d.Infof("ImportData: File  %s has been reopened after change detected ", d.NmonFile.CurFile)
		d.Infof("ImportData: Process pending Buffer before load data again ")
		//flush all existing chunks of data in buffer
		d.NmonFile.ProcessPending(points, d.TagMap)
		d.Infof("ImportData: Reset remaining  (%d) lines in buffer", len(d.NmonFile.PendingLines))
		//reset all remaining lines (from not completed chunks)
		d.NmonFile.ResetPending()
		d.NmonFile.ResetText()
		//if file has been rotated with format like /var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon
		//old file has been closed and a new one opened
		// we should now rescan definitions
		d.Infof("ImportData: File  %s should be rescanned for new sections/columns ", d.NmonFile.CurFile)
		pos, err := d.NmonFile.InitSectionDefs()
		if err != nil {
			d.Errorf("ImportData: Error on Section Initializations after reopen file :%s ", err)
			d.Errorf("ImportData: New file is not open. Reseting file on next interval")
			d.NmonFile = nil
			return err
		}

		d.Infof("ImportData: Updated File Info ID[%s] NAME[%s] FILE[%s] POS[%d]", d.cfg.ID, d.cfg.Name, d.NmonFile.CurFile, pos)
		// now last file has been closed and a new one created
		//PENDING delete from FileInfo last file
		db.AddOrUpdateNmonFileInfo(&config.NmonFileInfo{ID: d.cfg.ID, DeviceName: d.cfg.Name, FileName: d.NmonFile.CurFile, LastPosition: pos})

	}

	numlines, filepos := d.NmonFile.UpdateContent()
	if numlines > 0 {
		d.NmonFile.ProcessPending(points, d.TagMap)
		d.Infof("ImportData: Current File  Position is [%d] last processed Chunk %s ", filepos, d.NmonFile.LastTime.String())
		db.AddOrUpdateNmonFileInfo(&config.NmonFileInfo{ID: d.cfg.ID, DeviceName: d.cfg.Name, FileName: d.NmonFile.CurFile, LastPosition: filepos})
	}
	// Add last processed lines
	d.Infof("ImportData: this import has generated %d Datapoints", points.Length())
	return nil
}
