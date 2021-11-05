package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/timer"
	"kinger/common/config"
	"kinger/common/consts"
	_ "kinger/meta"
	"time"
	"kinger/gopuppy/common/evq"
)

var cService *chatService

type chatService struct {
	logic.LogicService
}

func (cs *chatService) Start(appID uint16) {
	cService = cs
	cs.OnStart(appID, consts.AppChat)
	config.LoadConfig()
	timer.StartTicks(500 * time.Millisecond)
	registerRpc()

	c := make(chan struct{})
	evq.CallLater(func() {
		loadBlackList()
		newChatMgr()
		cs.ReportRpcHandlers()
		close(c)
	})
	<- c
}

func (cs *chatService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		cs.ReportOnStop()
		close(c)
	})
	<- c
	cs.OnStop()
}

func main() {
	app.NewApplication(consts.AppChat, &chatService{}).Run()
}
