package reborn

import (
	"kinger/apps/game/module"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"math"
)

var mod *rebornModule

type rebornModule struct {

}

func (m *rebornModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("reborn")
	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("version", version)
		playerAttr.SetMapAttr("reborn", attr)
	}
	return &rebornComponent{attr: attr}
}

func (m *rebornModule) GetRebornCnt(player types.IPlayer) int {
	return player.GetComponent(consts.RebornCpt).(*rebornComponent).getRebornCnt()
}

func (m *rebornModule) GetRebornRemainDay(player types.IPlayer) int {
	return player.GetComponent(consts.RebornCpt).(*rebornComponent).getRebornRemainDay()
}

func onPlayerPvpLevelUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	pvpLevel := args[2].(int)
	player.GetComponent(consts.RebornCpt).(*rebornComponent).onPvpLevelUpdate(pvpLevel)
}

func onFixServer1Data(args ...interface{}) {
	player := args[0].(types.IPlayer)
	rebornCpt := player.GetComponent(consts.RebornCpt).(*rebornComponent)
	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	rebornCnt := rebornCpt.getRebornCnt()

	//var skybook int
	//if rebornCnt < 6 {
	//	rebornCntGameData := gamedata.GetGameData(consts.RebornCnt).(*gamedata.RebornCntGameData)
	//	for i := 1; i <= rebornCnt; i++ {
	//		skybook += rebornCntGameData.Cnt2BookAmount[i]
	//	}

	//	resCpt.ModifyResource(consts.SkyBook, skybook, true)
	//}

	privileges := rebornCpt.resetPrivileges()
	prestige := ( len(privileges) + int( math.Ceil(float64(resCpt.GetResource(consts.Prestige)) / 20000.0) ) ) * 20000
	if prestige > 0 {
		resCpt.SetResource(consts.Prestige, prestige)
	}

	glog.Infof("reborn begin fixServer1Data uid=%d, rebornCnt=%d, prestige=%d, privileges=%v",
		player.GetUid(), rebornCnt, prestige, privileges)
}

func Initialize() {
	mod = &rebornModule{}
	module.Reborn = mod
	registerRpc()
	eventhub.Subscribe(consts.EvPvpLevelUpdate, onPlayerPvpLevelUpdate)
	eventhub.Subscribe(consts.EvFixServer1Data, onFixServer1Data)
}
