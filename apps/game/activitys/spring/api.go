package spring

import (
	atypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/proto/pb"
	"kinger/common/consts"
	"time"
	"kinger/gopuppy/common/timer"
)

func Initialize() {
	mod = &activity{}
	mod.initActivityData()
}

func FetchActivityList(player types.IPlayer, aid int) (*pb.ActivityData, error) {
	p := newComponent(player)
	p.checkVersion(aid)
	rspData := &pb.ActivityData{ID: int32(aid)}
	detail := &pb.SpringHuodong{}
	p.forEachExchangeCnt(aid, func(goodsID, cnt int) {
		detail.ExchangeDatas = append(detail.ExchangeDatas, &pb.SpringExchangeData{
			GoodsID: int32(goodsID),
			ExchangeCnt: int32(cnt),
		})
	})
	rspData.Data, _ = detail.Marshal()
	return rspData, nil
}

func GetEventItemType() int {
	return consts.EventItem1
}

func OnGetTreasureRewardItemType(player types.IPlayer, itemAmount int) int {
	var itemType int
	cpt := atypes.IMod.NewPCM(player)
	p := newComponent(player)
	now := time.Now()
	atypes.IMod.ForEachActivityDataByType(consts.ActivityOfSpring, func(data atypes.IActivity) {
		aData := data.GetGameData()
		todayRewardItemAmount := p.getTreasureRewardItemAmount(aData.ID)
		if aData.ItemMaxDaily > 0 && todayRewardItemAmount >= aData.ItemMaxDaily {
			return
		}

		if cpt.ConformTime(aData.ID) && now.Before(data.GetTimeCondition().ItemEndTime) {
			itemType = consts.EventItem1
			p.setTreasureRewardItemAmount(aData.ID, todayRewardItemAmount + itemAmount)
		}
	})

	return itemType
}

func OnLogin(player types.IPlayer) {
	OnCrossDay(player, timer.GetDayNo())
}

func OnCrossDay(player types.IPlayer, curDayNo int) {
	if player.GetDataDayNo() == curDayNo {
		return
	}

	p := newComponent(player)
	atypes.IMod.ForEachActivityDataByType(consts.ActivityOfSpring, func(data atypes.IActivity) {
		aData := data.GetGameData()
		p.setTreasureRewardItemAmount(aData.ID, 0)
	})
}
