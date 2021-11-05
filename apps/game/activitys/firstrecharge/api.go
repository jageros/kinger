package firstrecharge

import (
	"kinger/gopuppy/common/eventhub"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
)

func AddEvent() {
	eventhub.Subscribe(consts.EvRecharge, onRecharge)
}

func Initialize() {
	mod = &activity{
		id2Reward: map[int]*gamedata.ActivityFirstRechargeRewardGameData{},
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
		resID, err := mod.getRewardFinshCondition(aid, rid)
		if err != nil {
			return nil, err
		}
		price := p.getRechargePrice(resID)
		rwcStr := resID + ":" + strconv.Itoa(price)
		finsh := func()int32{
			if p.hasBuy(aid, rid) {
				return 1
			}
			return 0
		}
		activityList.Activitys = append(activityList.Activitys, &pb.Activity{
			RewardID:        int32(rid),
			ReceiveStatus:   rst,
			RewardCondition: rwcStr,
			FinshNum:        finsh(),
			RewardList:      rwl,
		})
	}

	rspData.Data, _ = activityList.Marshal()
	return rspData, nil
}

func ReceiveReward(player types.IPlayer, aid, rid int, rd *pb.Reward) error {
	p := newComponent(player)
	canReceive := p.conformRewardCondition(aid, rid)
	if canReceive {
		err := p.giveReward(aid, rid, rd)
		if err != nil {
			return err
		}

		if cardAmount, ok := rd.RewardList["66"]; ok {
			cm := player.GetComponent(consts.CardCpt).(types.ICardComponent).GetCollectCard(66).GetAmount()
			if cm == int(cardAmount) {
				module.Televise.SendNotice(pb.TeleviseEnum_NewbieGiftGetCard, player.GetName(), uint32(66))
			}
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

func UpdateTabList(player types.IPlayer) {
	playerCom := newComponent(player)
	playerCom.updateActivityTagList()
}