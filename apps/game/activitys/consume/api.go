package consume

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"strconv"
)

func AddEvent() {
	//mod.addReloadCallbackRewardData()
	eventhub.Subscribe(consts.EvResUpdate, onConsume)
}

func Initialize() {
	mod = &activity{
		id2Reward: map[int]*gamedata.ActivityConsumeRewardGameData{},
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
		rwc, err := mod.getRewardFinshCondition(aid, rid)
		if err != nil {
			return nil, err
		}
		finsh := p.getTotalConsumeAmount(aid)
		rwcStr := strconv.Itoa(rwc)
		activityList.Activitys = append(activityList.Activitys, &pb.Activity{
			RewardID:        int32(rid),
			ReceiveStatus:   rst,
			RewardCondition: rwcStr,
			FinshNum:        int32(finsh),
			RewardList:      rwl,
		})
	}

	return rspData, nil
}

func ReceiveReward(player types.IPlayer, activityID, rewardID int, rd *pb.Reward) error {
	p := newComponent(player)
	canReceive := p.conformRewardCondition(activityID, rewardID)
	if canReceive {
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
