package main

import (
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/timer"
	_ "kinger/meta"
	"kinger/sdk"
	"time"
)

type sdkService struct {
	logic.LogicService
}

func (ss *sdkService) Start(appid uint16) {
	ss.OnStart(appid, consts.AppSdk)
	config.LoadConfig()

	timer.StartTicks(time.Second)

	c := make(chan struct{})
	evq.CallLater(func() {

		sdk.Initialize()
		initializeRouter()
		close(c)

	})

	<-c
}

func (ss *sdkService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		ss.ReportOnStop()
		close(c)
	})
	<-c
	ss.OnStop()
}

func main() {
	app.NewApplication(consts.AppSdk, &sdkService{}).Run()
}
