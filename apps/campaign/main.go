package main

import (
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/app"
	gconfig "kinger/gopuppy/common/config"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/rpubsub"
	"kinger/gopuppy/common/timer"
	_ "kinger/meta"
	"time"
)

var cService *campaignService

type campaignService struct {
	logic.LogicService
}

func (cs *campaignService) Start(appID uint16) {
	cService = cs
	cs.OnStart(appID, consts.AppCampaign)
	config.LoadConfig()

	rpubsub.Initialize(gconfig.GetRegionConfig().Redis.Addr)
	common.Init32UUidGenerator()
	timer.StartTicks(500 * time.Millisecond)

	c := make(chan struct{})
	evq.CallLater(func() {

		defer func() {
			err := recover()
			if err != nil {
				glog.TraceError("panic: %s", err)
				go func() {
					panic(err)
				}()
			}
		}()

		utils.RegisterDirtyWords()
		gamedata.Load()
		warMgr.initialize()
		campaignMgr.initialize()
		countryMgr.initialize()
		cityMgr.initialize()
		noticeMgr.initialize()
		playerMgr.initialize()
		sceneMgr.initialize()
		createCountryMgr.initialize()
		fieldMatchMgr.initialize()
		cityMatchMgr.initialize()
		warMgr.initializeTimer()
		registerRpc()
		cs.ReportRpcHandlers()
		close(c)
	})
	<-c
}

func (cs *campaignService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		cs.ReportOnStop()
		countryMgr.save(true)
		cityMgr.save(true)
		noticeMgr.save(true)
		sceneMgr.save()
		playerMgr.save(true)
		warMgr.save(true)
		close(c)
	})
	<-c
	cs.OnStop()
}

func main() {
	app.NewApplication(consts.AppCampaign, &campaignService{}).Run()
}
