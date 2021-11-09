package mission

import (
	"fmt"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	//"kinger/gopuppy/common/glog"
)

var (
	templates map[int]iMissionTemplate
	_         iMissionTemplate = &baseMissionTemplate{}
)

type iMissionTemplate interface {
	getData() *gamedata.Mission
	onInviteBattle() int
	onOpenTreasure() int
	onWxShare() int
	onWatchVideo() int
	onAddFriend() int
	onShareVideo() int
	onAccTreasure() int
	onPvpBattleEnd(fighterData *pb.EndFighterData, isWin bool) int
}

type baseMissionTemplate struct {
	data *gamedata.Mission
}

func (mt *baseMissionTemplate) getData() *gamedata.Mission {
	return mt.data
}

func (mt *baseMissionTemplate) onPvpBattleEnd(fighterData *pb.EndFighterData, isWin bool) int {
	return 0
}

func (mt *baseMissionTemplate) onInviteBattle() int {
	return 0
}

func (mt *baseMissionTemplate) onOpenTreasure() int {
	return 0
}

func (mt *baseMissionTemplate) onWxShare() int {
	return 0
}

func (mt *baseMissionTemplate) onAddFriend() int {
	return 0
}

func (mt *baseMissionTemplate) onShareVideo() int {
	return 0
}

func (mt *baseMissionTemplate) onWatchVideo() int {
	return 0
}

func (mt *baseMissionTemplate) onAccTreasure() int {
	return 0
}

type battleMissionTemplate struct {
	baseMissionTemplate
}

func (mt *battleMissionTemplate) onPvpBattleEnd(fighterData *pb.EndFighterData, isWin bool) int {
	//glog.Infof("battleMissionTemplate onPvpBattleEnd %d %d", mt.data.Camp, fighterData.Camp)
	if mt.data.Camp > 0 && mt.data.Camp != int(fighterData.Camp) {
		return 0
	}
	return 1
}

type battleWinMissionTemplate struct {
	baseMissionTemplate
}

func (mt *battleWinMissionTemplate) onPvpBattleEnd(fighterData *pb.EndFighterData, isWin bool) int {
	//glog.Infof("battleWinMissionTemplate onPvpBattleEnd %d %d %v", mt.data.Camp, fighterData.Camp, isWin)
	if !isWin || (mt.data.Camp > 0 && mt.data.Camp != int(fighterData.Camp)) {
		return 0
	}
	return 1
}

type useCardMissionTemplate struct {
	baseMissionTemplate
}

func (mt *useCardMissionTemplate) onPvpBattleEnd(fighterData *pb.EndFighterData, isWin bool) int {
	cnt := 0
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, gcardID := range fighterData.UseCards {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData == nil {
			continue
		}

		if mt.data.Camp > 0 && mt.data.Camp != cardData.Camp {
			continue
		}

		if mt.data.CardRare > 0 {
			if mt.data.CardRare == 4 {
				if cardData.Rare < 4 {
					continue
				}
			} else if mt.data.CardRare != cardData.Rare {
				continue
			}
		}

		cnt++
	}

	return cnt
}

type inviteBattleMissionTemplate struct {
	baseMissionTemplate
}

func (mt *inviteBattleMissionTemplate) onInviteBattle() int {
	return 1
}

type openTreasureMissionTemplate struct {
	baseMissionTemplate
}

func (mt *openTreasureMissionTemplate) onOpenTreasure() int {
	return 1
}

type wxShareMissionTemplate struct {
	baseMissionTemplate
}

func (mt *wxShareMissionTemplate) onWxShare() int {
	return 1
}

type addFriendMissionTemplate struct {
	baseMissionTemplate
}

func (mt *addFriendMissionTemplate) onAddFriend() int {
	return 1
}

type shareVideoMissionTemplate struct {
	baseMissionTemplate
}

func (mt *shareVideoMissionTemplate) onShareVideo() int {
	return 1
}

type watchVideoMissionTemplate struct {
	baseMissionTemplate
}

func (mt *watchVideoMissionTemplate) onWatchVideo() int {
	return 1
}

type accTreasureMissionTemplate struct {
	baseMissionTemplate
}

func (mt *accTreasureMissionTemplate) onAccTreasure() int {
	return 1
}

func doInitMissionTemplate(gdata gamedata.IGameData) {
	missionGameData := gdata.(*gamedata.MissionGameData)
	tpls := map[int]iMissionTemplate{}
	for missionID, mdata := range missionGameData.Missions {
		switch mdata.Type {
		case mtBattle:
			tpls[missionID] = &battleMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtBattleWin:
			tpls[missionID] = &battleWinMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtUseCard:
			tpls[missionID] = &useCardMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtInviteBattle:
			tpls[missionID] = &inviteBattleMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtOpenTreasure:
			tpls[missionID] = &openTreasureMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtWxShare:
			tpls[missionID] = &wxShareMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtAddFriend:
			tpls[missionID] = &addFriendMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtShareVideo:
			tpls[missionID] = &shareVideoMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtWatchVideo:
			tpls[missionID] = &watchVideoMissionTemplate{baseMissionTemplate{data: mdata}}
		case mtAccTreasure:
			tpls[missionID] = &accTreasureMissionTemplate{baseMissionTemplate{data: mdata}}
		default:
			panic(fmt.Sprintf("unknow mission type %d, id=%d", mdata.Type, missionID))
		}
	}

	templates = tpls
}

func initMissionTemplate() {
	missionGameData := gamedata.GetGameData(consts.Mission)
	missionGameData.AddReloadCallback(doInitMissionTemplate)
	doInitMissionTemplate(missionGameData)
}
