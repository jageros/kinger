package mission

import (
	"kinger/gopuppy/common/eventhub"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
	"kinger/common/consts"
	"kinger/proto/pb"
	"kinger/gamedata"
	"strconv"
	//"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"time"
)

var mod *missionModule

type missionModule struct {

}

func (m *missionModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("mission")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("mission", attr)
	}
	return &missionComponent{attr: attr}
}

func (m *missionModule) OnOpenTreasure(player types.IPlayer) {
	if player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).GetResource(consts.Score) <= 0 {
		return
	}
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onOpenTreasure()
	})
}

func (m *missionModule) OnInviteBattle(player types.IPlayer) {
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onInviteBattle()
	})
}

func (m *missionModule) OnWxShare(player types.IPlayer, shareTime int64) {
	if !utils.IsSameDay(shareTime, time.Now().Unix()) {
		return
	}

	if player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).GetResource(consts.Score) <= 0 {
		return
	}
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onWxShare()
	})
}

func (m *missionModule) OnWatchVideo(player types.IPlayer) {
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onWatchVideo()
	})
}

func (m *missionModule) OnAddFriend(player types.IPlayer) {
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onAddFriend()
	})
}

func (m *missionModule) OnShareVideo(player types.IPlayer) {
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onShareVideo()
	})
}

func (m *missionModule) OnAccTreasure(player types.IPlayer) {
	if player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).GetResource(consts.Score) <= 0 {
		return
	}
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onAccTreasure()
	})
}

func (m *missionModule) OnPvpBattleEnd(player types.IPlayer, fighterData *pb.EndFighterData, isWin bool) {
	player.GetComponent(consts.MissionCpt).(*missionComponent).tryAddMissionCnt(func(tpl iMissionTemplate) int {
		return tpl.onPvpBattleEnd(fighterData, isWin)
	})
}

func (m *missionModule) RefreashMission(player types.IPlayer) {
	if mc, ok := player.GetComponent(consts.MissionCpt).(*missionComponent); ok {
		mc.refreshMissionAfterNewbieSelectCamp()
	}
}

func (m *missionModule) GmAddMission(player types.IPlayer, args []string) error {
	if len(args) < 3 {
		return gamedata.GameError(1)
	}

	idx, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}
	if idx < 1 || idx > 3 {
		return gamedata.GameError(2)
	}

	missionID, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	tpl, ok := templates[missionID]
	if !ok {
		return gamedata.GameError(3)
	}

	mo := newMission(tpl.getData(), idx)
	if mo == nil {
		return gamedata.GameError(4)
	}

	player.GetComponent(consts.MissionCpt).(*missionComponent).gmAddMission(mo)
	return nil
}

func (m *missionModule) GmCompleteMission(player types.IPlayer) {
	player.GetComponent(consts.MissionCpt).(*missionComponent).gmCompleteMission()
}

func onBattleEnd(arg ... interface{}) {
	player, ok := arg[0].(types.IPlayer)
	if !ok {
		return
	}
	camp := int(arg[2].(*pb.EndFighterData).Camp)
	if p, ok := player.GetComponent(consts.MissionCpt).(*missionComponent); ok {
		num := p.getCampUseNum(camp)
		p.setCampUseNum(camp, num+1)
	}
}

func Initialize()  {
	mod = &missionModule{}
	module.Mission = mod
	initMissionTemplate()
	registerRpc()
	eventhub.Subscribe(consts.EvEndPvpBattle, onBattleEnd)
}
