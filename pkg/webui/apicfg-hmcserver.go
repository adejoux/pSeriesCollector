package webui

import (
	"time"

	"github.com/adejoux/pSeriesCollector/pkg/agent"
	"github.com/adejoux/pSeriesCollector/pkg/agent/devices/hmc"
	"github.com/adejoux/pSeriesCollector/pkg/config"
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
)

// NewAPICfgHMCServer HMCServer API REST creator
func NewAPICfgHMCServer(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/hmcserver", func() {
		m.Get("/", reqSignedIn, GetHMCServerServer)
		m.Post("/", reqSignedIn, bind(config.HMCCfg{}), AddHMCServerServer)
		m.Put("/:id", reqSignedIn, bind(config.HMCCfg{}), UpdateHMCServerServer)
		m.Delete("/:id", reqSignedIn, DeleteHMCServerServer)
		m.Get("/:id", reqSignedIn, GetHMCServerServerByID)
		m.Get("/checkondel/:id", reqSignedIn, GetHMCServerAffectOnDel)
		m.Post("/ping/", reqSignedIn, bind(config.HMCCfg{}), PingHMCServer)
	})

	return nil
}

// GetHMCServerServer Return Server Array
func GetHMCServerServer(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetHMCCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get HMCServer db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// AddHMCServerServer Insert new measurement groups to de internal BBDD --pending--
func AddHMCServerServer(ctx *Context, dev config.HMCCfg) {
	log.Printf("ADDING HMCServer Backend %+v", dev)
	affected, err := agent.MainConfig.Database.AddHMCCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateHMCServerServer --pending--
func UpdateHMCServerServer(ctx *Context, dev config.HMCCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateHMCCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update HMCServer db %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteHMCServerServer --pending--
func DeleteHMCServerServer(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelHMCCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetHMCServerServerByID --pending--
func GetHMCServerServerByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetHMCCfgByID(id)
	if err != nil {
		log.Warningf("Error on get HMCServer db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetHMCServerAffectOnDel --pending--
func GetHMCServerAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetHMCCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

//PingHMCServer Return ping result
func PingHMCServer(ctx *Context, cfg config.HMCCfg) {
	log.Infof("trying to ping influx server %s : %+v", cfg.ID, cfg)
	_, elapsed, message, err := hmc.Ping(&cfg)
	type result struct {
		Result  string
		Elapsed time.Duration
		Message string
	}
	if err != nil {
		log.Debugf("ERROR on ping HMCServerDB Server : %s", err)
		res := result{Result: "NOOK", Elapsed: elapsed, Message: err.Error()}
		ctx.JSON(400, res)
	} else {
		log.Debugf("OK on ping HMCServerDB Server %+v, %+v", elapsed, message)
		res := result{Result: "OK", Elapsed: elapsed, Message: message}
		ctx.JSON(200, res)
	}
}
