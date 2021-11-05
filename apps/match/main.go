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
	"kinger/gopuppy/common"
	"kinger/common/aicardpool"
)

var mService *matchService

type matchService struct {
	logic.LogicService
}

func (ms *matchService) Start(appID uint16) {
	mService = ms
	ms.OnStart(appID, consts.AppMatch)
	config.LoadConfig()
	timer.StartTicks(200 * time.Millisecond)
	rpubsub.Initialize(gconfig.GetRegionConfig().Redis.Addr)
	common.InitUUidGenerator("matchRobot")

	c := make(chan struct{})
	evq.CallLater(func() {
		gamedata.Load()
		gMatchMgr = newMatchMgr()
		robotMgr.loadRobot()
		gMatchMgr.beginHeartbeat()
		aicardpool.Load(consts.AppMatch, ms.AppID)
		registerRpc()
		ms.ReportRpcHandlers()
		close(c)
	})
	<- c
}

func (ms *matchService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		ms.ReportOnStop()
		robotMgr.saveRobot()
		aicardpool.Save()
		close(c)
	})
	<- c
	ms.OnStop()
}

func main() {
	app.NewApplication(consts.AppMatch, &matchService{}).Run()
}
