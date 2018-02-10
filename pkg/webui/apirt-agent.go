package webui

import (
	//	"github.com/go-macaron/binding"
	"github.com/adejoux/pSeriesCollector/pkg/agent"
	"gopkg.in/macaron.v1"
	//	"time"
)

// NewAPIRtAgent set API for the runtime management
func NewAPIRtAgent(m *macaron.Macaron) error {

	//	bind := binding.Bind

	m.Group("/api/rt/agent", func() {
		m.Get("/info/version/", reqSignedIn, RTGetVersion)
	})

	return nil
}

//RTGetVersion xx
func RTGetVersion(ctx *Context) {
	info := agent.GetRInfo()
	ctx.JSON(200, &info)
}
