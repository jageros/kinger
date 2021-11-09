package growplan

import (
	"fmt"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
)

func AddEvent() {
	eventhub.Subscribe(consts.EvEndPvpBattle, onFightEnd)
	eventhub.Subscribe(consts.EvResUpdate, onConsume)
	eventhub.Subscribe(consts.EvOpenTreasure, onOpenTreasure)
	eventhub.Subscribe(consts.EvCombat, onCombat)
	eventhub.Subscribe(consts.EvGetMissionReward, onFinshMission)
	eventhub.Subscribe(consts.EvShareBattleReport, onShareBattleReport)
	eventhub.Subscribe(consts.EVWatchBattleReport, onWatchBattleReport)
	eventhub.Subscribe(consts.EvMaxPvpLevelUpdate, onMaxPvpLevelUpdate)
	eventhub.Subscribe(consts.EvAddFriend, onAddFriend)
	eventhub.Subscribe(consts.EvCardUpdate, onCardUpdate)
	eventhub.Subscribe(consts.EvRecharge, onRecharge)
}

func Initialize() {
	mod = &activity{
		id2Reward: map[int]*gamedata.ActivityGrowPlanRewardGameData{},
	}
	mod.initActivityData()
	updateAllPlayerHint()
}

func FetchActivityList(player types.IPlayer, aid int) (*pb.ActivityData, error) {
	p := newComponent(player)
	rspData := &pb.ActivityData{}
	activityList := &pb.ActivityList{}
	rspData.ID = int32(aid)
	rewardIdList := mod.getRewardIdList(aid)

	for _, rid := range rewardIdList {
		rst := p.getRewardReceiveStatus(aid, rid)
		rwl := mod.getRewardMap(aid, rid)
		ty, rwc, err := mod.getRewardFinshCondition(aid, rid)
		if err != nil {
			return nil, err
		}
		gid := mod.getGoodsId(aid, rid)
		if p.isBuyGift(aid, rid) {
			gid = "0"
		}
		rwcStr := fmt.Sprintf("%d:%d:%d:%s", ty, rwc[0], rwc[1], gid)
		finsh := p.getFinshNum(aid, rid)
		activityList.Activitys = append(activityList.Activitys, &pb.Activity{
			RewardID:        int32(rid),
			ReceiveStatus:   rst,
			RewardCondition: rwcStr,
			FinshNum:        int32(finsh),
			RewardList:      rwl,
		})
	}

	rspData.Data, _ = activityList.Marshal()
	return rspData, nil
}

func ReceiveReward(player types.IPlayer, activityID, rewardID int, rd *pb.Reward) error {
	p := newComponent(player)
	canReceive := p.getRewardReceiveStatus(activityID, rewardID)
	if canReceive == pb.ActivityReceiveStatus_CanReceive {
		err := p.giveReward(activityID, rewardID, rd)
		if err != nil {
			return err
		}
		return nil
	}
	err := gamedata.GameError(aTypes.CanNotReceiveRewardError)
	return err
}

func OnLogin(player types.IPlayer) {
	p := newComponent(player)
	p.OnCrossDays(timer.GetDayNo())
	p.updateHint()
}

func UpdateTabList(player types.IPlayer) {
	p := newComponent(player)
	p.updateActivityTagList()
}

func OnCrossDay(player types.IPlayer, dayno int) {
	p := newComponent(player)
	p.OnCrossDays(dayno)
}
