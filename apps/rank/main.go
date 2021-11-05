package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/app"
	gconfig "kinger/gopuppy/common/config"
	"kinger/gopuppy/common/rpubsub"
	"kinger/gopuppy/common/timer"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	_ "kinger/meta"
	"time"
	"kinger/gopuppy/common/evq"
)

var rService *rankService

type rankService struct {
	logic.LogicService
}

func (rs *rankService) Start(appID uint16) {
	rService = rs
	rs.OnStart(appID, consts.AppRank)
	config.LoadConfig()
	timer.StartTicks(500 * time.Millisecond)
	rpubsub.Initialize(gconfig.GetRegionConfig().Redis.Addr)

	c := make(chan struct{})
	evq.CallLater(func() {
		gamedata.Load()
		newRankMgr()
		rankMgr.loadBoard()
		registerRpc()
		rs.ReportRpcHandlers()
		close(c)
	})
	<- c
}

func (rs *rankService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		rs.ReportOnStop()
		rankMgr.refreshCurRankList(true)
		rankMgr.save()
		close(c)
	})
	<- c
	rs.OnStop()
}

func main() {
	app.NewApplication(consts.AppRank, &rankService{}).Run()
}
