package nmon

import (
	"github.com/adejoux/pSeriesCollector/pkg/data/pointarray"
)

// ImportData getmon N data from Remote devices
func (d *Server) ImportData(points *pointarray.PointArray) error {

	d.Infof("Import Nmon data on remote device (%s) ", d.cfg.NmonIP)
	if d.nmonFile == nil {
		d.Infof("Initializing Nmon Remote File")
		d.nmonFile = NewNmonFile(d.client, d.GetLogger(), d.cfg.NmonFilePath, d.cfg.Name)
		d.nmonFile.Init()
		d.Debugf("Found Dataseries: %#+v", d.nmonFile.DataSeries)
		d.Debugf("Found Content %s", d.nmonFile.TextContent)

	}

	d.nmonFile.UpdateContent()
	// Add last processed lines

	d.nmonFile.ProcessPending(points, d.TagMap)

	d.Debugf("SFTP status %#+v", d.client)

	return nil
}
