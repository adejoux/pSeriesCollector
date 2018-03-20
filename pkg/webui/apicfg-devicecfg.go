package webui

import (
	//	"time"

	"time"

	"github.com/adejoux/pSeriesCollector/pkg/agent"
	"github.com/adejoux/pSeriesCollector/pkg/agent/devices/nmon"
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
)

// NewAPICfgDevice DeviceCfg API REST creator
func NewAPICfgDevice(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/devices", func() {
		m.Get("/", reqSignedIn, GetDeviceCfg)
		m.Post("/", reqSignedIn, bind(config.DeviceCfg{}), AddDeviceCfg)
		m.Put("/:id", reqSignedIn, bind(config.DeviceCfg{}), UpdateDeviceCfg)
		m.Delete("/:id", reqSignedIn, DeleteDeviceCfg)
		m.Get("/:id", reqSignedIn, GetDeviceCfgByID)
		m.Get("/checkondel/:id", reqSignedIn, GetDeviceCfgAffectOnDel)
		m.Post("/ping/", reqSignedIn, bind(config.DeviceCfg{}), PingDeviceCfg)
	})

	return nil
}

// GetDeviceCfg Return Server Array
func GetDeviceCfg(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetDeviceCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Device :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// AddDeviceCfg Insert new measurement groups to de internal BBDD --pending--
func AddDeviceCfg(ctx *Context, dev config.DeviceCfg) {
	log.Printf("ADDING Device %+v", dev)
	affected, err := agent.MainConfig.Database.AddDeviceCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateDeviceCfg --pending--
func UpdateDeviceCfg(ctx *Context, dev config.DeviceCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateDeviceCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update device %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteDeviceCfg --pending--
func DeleteDeviceCfg(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelDeviceCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetDeviceCfgByID --pending--
func GetDeviceCfgByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetDeviceCfgByID(id)
	if err != nil {
		log.Warningf("Error on get device db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// GetDeviceCfgAffectOnDel --pending--
func GetDeviceCfgAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetDeviceCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

//PingDeviceCfg Return ping result
func PingDeviceCfg(ctx *Context, cfg config.DeviceCfg) {
	log.Infof("trying to ping influx server %s : %+v", cfg.ID, cfg)
	_, elapsed, message, err := nmon.Ping(&cfg, log, false, "")
	type result struct {
		Result  string
		Elapsed time.Duration
		Message string
	}
	if err != nil {
		log.Debugf("ERROR on ping Device : %s", err)
		res := result{Result: "NOOK", Elapsed: elapsed, Message: err.Error()}
		ctx.JSON(400, res)
	} else {
		log.Debugf("OK on ping Device Server %+v, %+v", elapsed, message)
		res := result{Result: "OK", Elapsed: elapsed, Message: message}
		ctx.JSON(200, res)
	}
}
