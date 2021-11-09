package main

import (
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/app"
	gconfig "kinger/gopuppy/common/config"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/rpubsub"
	"kinger/gopuppy/common/timer"
	_ "kinger/meta"
	"time"
)

var vService *videoService

type videoService struct {
	logic.LogicService
}

func (vs *videoService) Start(appID uint16) {
	vService = vs
	vs.OnStart(appID, consts.AppVideo)
	config.LoadConfig()
	timer.StartTicks(500 * time.Millisecond)
	rpubsub.Initialize(gconfig.GetRegionConfig().Redis.Addr)

	c := make(chan struct{})
	evq.CallLater(func() {
		gamedata.Load()
		newVideoMgr()
		registerRpc()
		vs.ReportRpcHandlers()
		close(c)
	})
	<-c
}

func (vs *videoService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		vs.ReportOnStop()
		videoMgr.save()
		close(c)
	})
	<-c
	vs.OnStop()
}

func main() {
	app.NewApplication(consts.AppVideo, &videoService{}).Run()
}
