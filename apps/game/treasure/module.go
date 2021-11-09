package treasure

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
)

var mod *treasureModule

type treasureModule struct {
}

func (t *treasureModule) NewTreasureComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	treasureAttr := playerAttr.GetMapAttr("treasure")
	if treasureAttr == nil {
		treasureAttr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("treasure", treasureAttr)
	}
	return &treasureComponent{attr: treasureAttr}
}

func (t *treasureModule) OpenTreasureByModelID(player types.IPlayer, modelID string, isDobule bool) *pb.OpenTreasureReply {
	return player.GetComponent(consts.TreasureCpt).(*treasureComponent).OpenTreasureByModelID(modelID, isDobule)
}

func (t *treasureModule) GetDayAccTicketCanAdd(player types.IPlayer) int {
	if config.GetConfig().IsXfServer() {
		return 0
	}
	if player.IsVip() {
		if config.GetConfig().HostID >= 1001 {
			return module.OutStatus.BuffAccTreasureCnt(player, 0)
		}

		if player.IsForeverVip() {
			return module.OutStatus.BuffAccTreasureCnt(player, 20)
		} else {
			return module.OutStatus.BuffAccTreasureCnt(player, 10)
		}
	} else {
		return module.OutStatus.BuffAccTreasureCnt(player, 3)
	}
}

func (t *treasureModule) WxHelpDoubleDailyTreasure(player types.IPlayer, treasureID uint32, helperUid common.UUid, helperHeadImg,
	helperHeadFrame, helperName string) bool {
	return player.GetComponent(consts.TreasureCpt).(*treasureComponent).wxHelpDoubleDailyTreasure(treasureID, helperUid,
		helperHeadImg, helperHeadFrame, helperName)
}

func nextDay() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		if !player.IsRobot() {
			treasureCpt := player.GetComponent(consts.TreasureCpt).(*treasureComponent)
			treasureCpt.AddDailyTreasure(false)
		}
	})
}

func onReborn(args ...interface{}) {
	player := args[0].(types.IPlayer)
	player.GetComponent(consts.TreasureCpt).(*treasureComponent).onReborn()
}

func Initialize() {
	mod = &treasureModule{}
	module.Treasure = mod
	registerRpc()
	timer.RunEveryDay(0, 0, 0, nextDay)
	eventhub.Subscribe(consts.EvReborn, onReborn)
}
