package nmon

import (
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
)

// ImportData getmon N data from Remote devices
func (d *Server) ImportData(points *pointarray.PointArray) error {

	d.Infof("Import Nmon data on remote device (%s) ", d.cfg.NmonIP)
	if d.nmonFile == nil {
		d.Infof("Initializing Nmon Remote File")
		nf := NewNmonFile(d.client, d.GetLogger(), d.cfg.NmonFilePath, d.cfg.Name)
		filepos, err := nf.Init()
		if err != nil {
			d.Errorf("Something happen on Initialize Nmon file: %s [Reopen Again....]", err)
			return err
		}
		d.nmonFile = nf
		// Got last known position
		info, err := db.GetNmonFileInfoByIDFile(d.cfg.ID, d.nmonFile.CurFile)
		if err != nil {
			d.Debugf("Warning on get file info for ID [%s] and file [%s] ", d.cfg.ID, d.nmonFile.CurFile)
			d.Infof("Current File Position %s is: %d", d.nmonFile.CurFile, filepos)
		} else {
			d.nmonFile.SetPosition(info.LastPosition)
			d.Infof("Updated File Position %s now to: %d", d.nmonFile.CurFile, info.LastPosition)
		}
		d.Debugf("Found Dataseries: %#+v", d.nmonFile.DataSeries)
		d.Debugf("Found Content %s", d.nmonFile.TextContent)
	}

	if d.nmonFile.ReopenIfChanged() {
		//flush all existing chunks of data in buffer
		d.nmonFile.ProcessPending(points, d.TagMap)
		//reset all remaining lines (from not completed chunks)
		d.nmonFile.ResetPending()
		//if file has been rotated with format like /var/log/nmon/%{hostname}_%Y%m%d_%H%M.nmon
		//old file has been closed and a new one opened
		// we should now rescan definitions
		d.Infof("File  %s should be rescanned for new sections/columns ", d.nmonFile.CurFile)
		pos, err := d.nmonFile.InitSectionDefs()
		if err != nil {
			d.Errorf("Error on Section Initializations after reopen file :%s ", err)
			return err
		}

		// now last file has been closed and a new one created
		//PENDING delete from FileInfo last file
		db.AddOrUpdateNmonFileInfo(&config.NmonFileInfo{ID: d.cfg.ID, DeviceName: d.cfg.Name, FileName: d.nmonFile.CurFile, LastPosition: pos})

	}

	numlines, filepos := d.nmonFile.UpdateContent()
	if numlines > 0 {
		d.nmonFile.ProcessPending(points, d.TagMap)
		d.Infof("Current File  Position is [%d] last processed Chunk %s ", filepos, d.nmonFile.LastTime.String())
		db.AddOrUpdateNmonFileInfo(&config.NmonFileInfo{ID: d.cfg.ID, DeviceName: d.cfg.Name, FileName: d.nmonFile.CurFile, LastPosition: filepos})
	}
	// Add last processed lines
	return nil
}
