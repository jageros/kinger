package online

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"kinger/proto/pb"
)

func AddEvent() {
	//timer.AddTicker(time.Second*3, updateAllPlayerHint)
}

func Initialize() {
	mod = &activity{
		id2Reward: map[int]*gamedata.ActivityOnlineRewardGameData{},
	}
	mod.initActivityData()
	updateAllPlayerInfo()
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
		finsh := 1
		if rst == pb.ActivityReceiveStatus_CanNotReceive {
			finsh = 0
		}
		activityList.Activitys = append(activityList.Activitys, &pb.Activity{
			RewardID:        int32(rid),
			ReceiveStatus:   rst,
			RewardCondition: rwc,
			FinshNum:        int32(finsh),
			RewardList:      rwl,
		})
	}

	rspData.Data, _ = activityList.Marshal()
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
	playerCom := newComponent(player)
	playerCom.updateHint()
}

func UpdataPlayerInfo(player types.IPlayer) {
	p := newComponent(player)
	p.updateHint()
	p.updateReceiveStatus()
}