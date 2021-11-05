package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/rpubsub"
	"kinger/gopuppy/common/timer"
	kconfig "kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	_ "kinger/meta"
	"time"
	"kinger/gopuppy/common/evq"
)

var bService *battleService

type battleService struct {
	logic.LogicService
	poolGameData *gamedata.PoolGameData
}

func (bs *battleService) Start(appid uint16) {
	bService = bs
	bs.OnStart(appid, consts.AppBattle)
	kconfig.LoadConfig()
	common.InitUUidGenerator("battle")
	timer.StartTicks(500 * time.Millisecond)
	rpubsub.Initialize(config.GetRegionConfig().Redis.Addr)

	c := make(chan struct{})
	evq.CallLater(func() {
		defer func() {
			err := recover()
			if err != nil {
				go func() {
					panic(err)
				}()
			}
		}()

		gamedata.Load()
		initSkillTarget()
		initSkill()
		initBonus()
		registerRpc()
		bs.ReportRpcHandlers()
		close(c)
	})
	<-c
}

func (bs *battleService) Stop() {
	c := make(chan struct{})
	evq.CallLater(func() {
		bs.ReportOnStop()
		for _, b := range mgr.id2Battle {
			b.save(true)
		}
		close(c)
	})
	<-c
	bs.OnStop()
}

func (bs *battleService) getPoolGameData() *gamedata.PoolGameData {
	if bs.poolGameData == nil {
		bs.poolGameData = gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	}
	return bs.poolGameData
}

func main() {
	app.NewApplication(consts.AppBattle, &battleService{}).Run()
}
