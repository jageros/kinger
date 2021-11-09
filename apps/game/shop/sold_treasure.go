package shop

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"strconv"
	//"time"
	"kinger/gopuppy/common/timer"
)

// 军备宝箱
type iSoldTreasure interface {
	packMsg() *pb.SoldTreasureData
	buy(treasureModelID string) (*pb.BuySoldTreasureReply, error)
	onLogout()
	onCrossDay()
	onPvpLevelUpdate()
}

type nilSoldTreasureSt struct {
}

func (st *nilSoldTreasureSt) packMsg() *pb.SoldTreasureData { return nil }
func (st *nilSoldTreasureSt) buy(treasureModelID string) (*pb.BuySoldTreasureReply, error) {
	return nil, gamedata.InternalErr
}
func (st *nilSoldTreasureSt) onLogout()         {}
func (st *nilSoldTreasureSt) onCrossDay()       {}
func (st *nilSoldTreasureSt) onPvpLevelUpdate() {}

type soldTreasureSt struct {
	attr        *attribute.MapAttr
	campIdxAttr *attribute.MapAttr
	player      types.IPlayer
	team        int
	//canBuyTimer *timer.Timer
}

var nilSoldTreasure = &nilSoldTreasureSt{}

func newSoldTreasure(player types.IPlayer, cptAttr *attribute.MapAttr) iSoldTreasure {
	if !config.GetConfig().IsXfServer() {
		return nilSoldTreasure
	}

	attr := cptAttr.GetMapAttr("sTreasure")
	if attr == nil {
		attr = attribute.NewMapAttr()
		cptAttr.SetMapAttr("sTreasure", attr)
	}

	campIdxAttr := attr.GetMapAttr("campIdx")
	if campIdxAttr == nil {
		campIdxAttr = attribute.NewMapAttr()
		attr.SetMapAttr("campIdx", campIdxAttr)
	}

	st := &soldTreasureSt{
		player:      player,
		attr:        attr,
		campIdxAttr: campIdxAttr,
		team:        player.GetPvpTeam(),
	}
	//st.beginCanBuyTimer()
	return st
}

/*
func (st *soldTreasureSt) beginCanBuyTimer() {
	st.stopCanBuyTimer()
	remainTime := st.getRemainTime()
	if remainTime > 0 {
		st.canBuyTimer = timer.AfterFunc(time.Duration(remainTime) * time.Second, func() {
			for _, camp := range []int{consts.Wei, consts.Shu, consts.Wu} {
				st.setCampIdx(camp, 0)
			}
			st.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_SoldTreasure)
		})
	}
}

func (st *soldTreasureSt) stopCanBuyTimer() {
	if st.canBuyTimer != nil {
		st.canBuyTimer.Cancel()
		st.canBuyTimer = nil
	}
}
*/

func (st *soldTreasureSt) onLogout() {
	//st.stopCanBuyTimer()
}

func (st *soldTreasureSt) onCrossDay() {
	for _, camp := range []int{consts.Wei, consts.Shu, consts.Wu} {
		st.setCampIdx(camp, 0)
	}
	st.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_SoldTreasure)
}

func (st *soldTreasureSt) onPvpLevelUpdate() {
	team := st.player.GetPvpTeam()
	if st.team == team {
		return
	}
	st.team = team
	st.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_SoldTreasure)
}

/*
func (st *soldTreasureSt) getRemainTime() int {
	timeout := st.attr.GetInt64("timeout")
	remainTime := timeout - time.Now().Unix()
	if remainTime < 0 {
		remainTime = 0
	}
	return int(remainTime)
}
*/

func (st *soldTreasureSt) getCampIdx(camp int) int {
	return st.campIdxAttr.GetInt(strconv.Itoa(camp))
}

func (st *soldTreasureSt) setCampIdx(camp, idx int) {
	st.campIdxAttr.SetInt(strconv.Itoa(camp), idx)
}

func (st *soldTreasureSt) getNextCanBuyTreasure(camp int) *gamedata.SoldTreasure {
	return gamedata.GetSoldTreasureGameData().GetCampTreasure(st.player.GetArea(), camp, st.player.GetPvpTeam(),
		st.getCampIdx(camp))
}

func (st *soldTreasureSt) getLastCanBuyTreasure(camp int) *gamedata.SoldTreasure {
	gdata := gamedata.GetSoldTreasureGameData()
	idx := st.getCampIdx(camp)
	area := st.player.GetArea()
	for i := idx - 1; i >= 0; i-- {
		t := gdata.GetCampTreasure(area, camp, st.player.GetPvpTeam(), i)
		if t != nil {
			return t
		}
	}
	return nil
}

func (st *soldTreasureSt) packMsg() *pb.SoldTreasureData {
	msg := &pb.SoldTreasureData{}

	for _, camp := range []int{consts.Wei, consts.Shu, consts.Wu} {
		t := st.getNextCanBuyTreasure(camp)
		var remainTime int32
		if t == nil {
			t = st.getLastCanBuyTreasure(camp)
			remainTime = int32(timer.TimeDelta(0, 0, 0).Seconds())
		}

		if t != nil {
			msg.SoldTreasures = append(msg.SoldTreasures, &pb.SoldTreasure{
				TreasureModelID: t.TreasureModelID,
				NeedJade:        int32(t.JadePrice),
				NexRemainTime:   remainTime,
			})
		}
	}

	return msg
}

func (st *soldTreasureSt) buy(treasureModelID string) (*pb.BuySoldTreasureReply, error) {

	soldTreasureGameData := gamedata.GetSoldTreasureGameData()
	soldTreasure := soldTreasureGameData.GetTreasureByID(treasureModelID)
	if soldTreasure == nil {
		return nil, gamedata.GameError(1)
	}

	soldTreasure2 := st.getNextCanBuyTreasure(soldTreasure.Camp)
	if soldTreasure2 == nil || soldTreasure2.TreasureModelID != soldTreasure.TreasureModelID {
		return nil, gamedata.GameError(2)
	}

	var resType, price int
	if soldTreasure.BowlderPrice > 0 {

		resType = consts.Bowlder
		price = soldTreasure.BowlderPrice
		if price > 0 {
			if !st.player.HasBowlder(price) {
				return nil, gamedata.GameError(4)
			}
			st.player.SubBowlder(price, consts.RmrSoldTreasure)
		}

	} else {

		resType = consts.Jade
		price = soldTreasure.JadePrice
		if price > 0 {
			resCpt := st.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
			if !resCpt.HasResource(consts.Jade, price) {
				return nil, gamedata.GameError(3)
			}
			resCpt.ModifyResource(consts.Jade, -price, consts.RmrSoldTreasure)
		}

	}

	treasureReward := st.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
		treasureModelID, false)

	st.setCampIdx(soldTreasure.Camp, st.getCampIdx(soldTreasure.Camp)+1)
	//if st.getNextCanBuyTreasure(soldTreasure.Camp) == nil {
	//remainTime := timer.TimeDelta(0, 0, 0)
	//st.attr.SetInt64("timeout", time.Now().Add(remainTime).Unix())
	//st.beginCanBuyTimer()
	//}

	mod.LogShopBuyItem(st.player, fmt.Sprintf("soldTreasure_%s", treasureModelID),
		fmt.Sprintf("军备宝箱_%s", treasureModelID), 1, "shop",
		strconv.Itoa(resType), module.Player.GetResourceName(resType), price, "")

	return &pb.BuySoldTreasureReply{
		TreasureReward:  treasureReward,
		NewSoldTreasure: st.packMsg(),
	}, nil
}
