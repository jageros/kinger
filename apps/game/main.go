package main

import (
	"kinger/apps/game/activitys"
	"kinger/apps/game/rank"
	"kinger/apps/game/televise"
	"time"

	"kinger/gopuppy/common"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/timer"
	//"kinger/gopuppy/network"
	//"kinger/apps/game/campaign"
	"kinger/apps/game/bag"
	"kinger/apps/game/campaign"
	"kinger/apps/game/cardpool"
	"kinger/apps/game/giftcode"
	"kinger/apps/game/huodong"
	"kinger/apps/game/level"
	"kinger/apps/game/mail"
	"kinger/apps/game/mission"
	"kinger/apps/game/module"
	"kinger/apps/game/outstatus"
	"kinger/apps/game/player"
	"kinger/apps/game/pvp"
	"kinger/apps/game/reborn"
	"kinger/apps/game/reward"
	"kinger/apps/game/shop"
	"kinger/apps/game/social"
	"kinger/apps/game/treasure"
	"kinger/apps/game/tutorial"
	"kinger/apps/game/web"
	"kinger/apps/game/wxgame"
	"kinger/common/aicardpool"
	kconfig "kinger/common/config"
	kconsts "kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/rpubsub"
	_ "kinger/meta"
	"kinger/sdk"
)

type gameService struct {
	logic.LogicService
	//peer *network.Peer
}

func (gs *gameService) GetAppID() uint32 {
	return gs.AppID
}

func (gs *gameService) GetRegion() uint32 {
	return gs.Region
}

func (gs *gameService) Start(appid uint16) {
	module.Service = gs
	gs.OnStart(appid, consts.AppGame)
	kconfig.LoadConfig()

	rpubsub.Initialize(config.GetRegionConfig().Redis.Addr)
	//common.InitUUidGenerator("player")
	common.InitUUidGenerator("battle")
	common.InitUUidGenerator("tourist")
	common.Init32UUidGenerator()
	timer.StartTicks(500 * time.Millisecond)

	c := make(chan struct{})
	evq.CallLater(func() {

		rpubsub.Subscribe("reload_config", func(i map[string]interface{}) {
			err := kconfig.ReloadConfig()
			if err == nil {
				evq.PostEvent(evq.NewCommonEvent(kconsts.EvReloadConfig))
				glog.Infof("reload_config ok")
			} else {
				glog.Errorf("reload_config err %s", err)
			}
		})

		utils.RegisterDirtyWords()
		utils.InitForbidList()
		gamedata.Load()
		reward.Initialize()
		player.Initialize()
		level.Initialize()
		cardpool.Initialize()
		pvp.Initialize()
		treasure.Initialize()
		tutorial.Initialize()
		giftcode.Initialize()
		social.Initialize()
		wxgame.Initialize()
		shop.Initialize()
		mission.Initialize()
		mail.Initialize()
		web.Initialize()
		sdk.Initialize()
		huodong.Initialize()
		bag.Initialize()
		reborn.Initialize()
		outstatus.Initialize()
		campaign.Initialize()
		activitys.Initialize()
		televise.Initialize()
		rank.Initialize()
		aicardpool.Load(consts.AppGame, gs.AppID)

		gs.ReportRpcHandlers()
		close(c)

	})

	<-c
}

func (gs *gameService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		gs.ReportOnStop()
		player.OnServerStop()
		cardpool.SaveAllCardLog()
		bag.OnServerStop()
		aicardpool.Save()
		close(c)
	})
	<-c
	gs.OnStop()
}

func main() {
	app.NewApplication(consts.AppGame, &gameService{}).Run()
}
