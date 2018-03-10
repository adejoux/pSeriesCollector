package pointarray

import (
	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

// PointArray new array of influx points
type PointArray struct {
	//Internal stats
	MetSent   int64
	MetError  int64
	MeasSent  int64
	MeasError int64
	//influx points
	PtArray []*client.Point
	log     *logrus.Logger
	bpts    *client.BatchPoints
}

// NewPointArray create a point array struct
func New(l *logrus.Logger, bpts *client.BatchPoints) *PointArray {
	return &PointArray{0, 0, 0, 0, nil, l, bpts}
}

// Append add influx points to the array
func (pa *PointArray) Append(meas string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	p, err := client.NewPoint(meas, tags, fields, t)
	if err != nil {
		pa.log.Warnf("error in influx point building:%s", err)
		pa.MeasError++
		pa.MetError += int64(len(fields))
	} else {
		pa.log.Debugf("GENERATED INFLUX POINT[%s] value: %+v", meas, p)
		pa.PtArray = append(pa.PtArray, p)
		pa.MeasSent++
		pa.MetSent += int64(len(fields))
	}
}

// Length  return num of attached points
func (pa *PointArray) Length() int {
	return len(pa.PtArray)
}

// Flush set dat
func (pa *PointArray) Flush() {
	if pa.bpts != nil {
		(*pa.bpts).AddPoints(pa.PtArray)
	}
}
