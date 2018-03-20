package impexp

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/adejoux/pSeriesCollector/pkg/agent"
	"github.com/adejoux/pSeriesCollector/pkg/config"
)

var (
	log     *logrus.Logger
	confDir string              //Needed to get File Filters data
	dbc     *config.DatabaseCfg //Needed to get Custom Filter  data
)

// SetConfDir  enable load File Filters from anywhere in the our FS.
func SetConfDir(dir string) {
	confDir = dir
}

// SetDB load database config to load data if needed (used in filters)
func SetDB(db *config.DatabaseCfg) {
	dbc = db
}

// SetLogger set output log
func SetLogger(l *logrus.Logger) {
	log = l
}

// ExportInfo Main export Data type
type ExportInfo struct {
	FileName      string
	Description   string
	Author        string
	Tags          string
	AgentVersion  string
	ExportVersion string
	CreationDate  time.Time
}

// EIOptions export/import options
type EIOptions struct {
	Recursive   bool   //Export Option
	AutoRename  bool   //Import Option
	AlternateID string //Import Option
}

// ExportObject Base type for any object to export
type ExportObject struct {
	ObjectTypeID string
	ObjectID     string
	Options      *EIOptions
	ObjectCfg    interface{}
	Error        string
}

// ExportData the runtime measurement config
type ExportData struct {
	Info       *ExportInfo
	Objects    []*ExportObject
	tmpObjects []*ExportObject //only for temporal use
}

// NewExport ExportData type creator
func NewExport(info *ExportInfo) *ExportData {
	if len(agent.Version) > 0 {
		info.AgentVersion = agent.Version
	} else {
		info.AgentVersion = "debug"
	}

	info.ExportVersion = "1.0"
	info.CreationDate = time.Now()
	return &ExportData{
		Info: info,
	}
}

func checkIfExistOnArray(list []*ExportObject, ObjType string, id string) bool {
	for _, v := range list {
		if v.ObjectTypeID == ObjType && v.ObjectID == id {
			return true
		}
	}
	return false
}

// PrependObject prepend a new object to the ExportData type
func (e *ExportData) PrependObject(obj *ExportObject) {
	if checkIfExistOnArray(e.Objects, obj.ObjectTypeID, obj.ObjectID) == true {
		return
	}
	e.tmpObjects = append([]*ExportObject{obj}, e.tmpObjects...)
}

// UpdateTmpObject update temporaty object
func (e *ExportData) UpdateTmpObject() {
	//we need remove duplicated objects on the auxiliar array
	objectList := []*ExportObject{}
	for i := 0; i < len(e.tmpObjects); i++ {
		v := e.tmpObjects[i]
		if checkIfExistOnArray(objectList, v.ObjectTypeID, v.ObjectID) == false {
			objectList = append(objectList, v)
		}
	}
	e.Objects = append(e.Objects, objectList...)
	e.tmpObjects = nil
}

// Export  exports data
func (e *ExportData) Export(ObjType string, id string, recursive bool, level int) error {
	switch ObjType {
	case "hmcservercfg":
		//contains sensible data
		v, err := dbc.GetHMCCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "hmcservercfg", ObjectID: id, ObjectCfg: v})
		if !recursive {
			break
		}
		e.Export("influxcfg", v.OutDB, recursive, level+1)
	case "devicecfg":
		//contains sensible data
		v, err := dbc.GetDeviceCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "devicecfg", ObjectID: id, ObjectCfg: v})
		if !recursive {
			break
		}
		e.Export("influxcfg", v.NmonOutDB, recursive, level+1)
	case "influxcfg":
		v, err := dbc.GetInfluxCfgByID(id)
		if err != nil {
			return err
		}
		e.PrependObject(&ExportObject{ObjectTypeID: "influxcfg", ObjectID: id, ObjectCfg: v})
	default:
		return fmt.Errorf("Unknown type object type %s ", ObjType)
	}
	if level == 0 {
		e.UpdateTmpObject()
	}
	return nil
}
