package spring

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	//"kinger/gopuppy/common/timer"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"strconv"
)

type activityCom struct {
	ipc    aTypes.IPlayerCom
	player types.IPlayer
	attr   *attribute.MapAttr
}

func newComponent(player types.IPlayer) *activityCom {
	pcm := aTypes.IMod.NewPCM(player)
	return &activityCom{
		ipc:    pcm,
		player: player,
		attr:   pcm.InitAttr(aTypes.SpringActivity),
	}
}

func (c *activityCom) getActivityAttr(aid int) *attribute.MapAttr {
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	return attr
}

func (c *activityCom) getExchangeCntAttr(aid int) *attribute.MapAttr {
	return c.getActivityAttr(aid).GetMapAttr("exchangeCnt")
}

func (c *activityCom) getTreasureRewardItemAmount(aid int) int {
	return c.getActivityAttr(aid).GetInt("rewardItem")
}

func (c *activityCom) setTreasureRewardItemAmount(aid int, amount int) {
	c.getActivityAttr(aid).SetInt("rewardItem", amount)
}

func (c *activityCom) getExchangeCnt(aid, goodsID int) int {
	exchangeCntAttr := c.getExchangeCntAttr(aid)
	if exchangeCntAttr == nil {
		return 0
	}
	return exchangeCntAttr.GetInt(strconv.Itoa(goodsID))
}

func (c *activityCom) onExchangeGoods(aid int, goodsData *gamedata.HuodongGoods) {
	if goodsData.ExchangeCnt > 0 {
		exchangeCntAttr := c.getExchangeCntAttr(aid)
		if exchangeCntAttr == nil {
			exchangeCntAttr = attribute.NewMapAttr()
			c.getActivityAttr(aid).SetMapAttr("exchangeCnt", exchangeCntAttr)
		}

		key := strconv.Itoa(goodsData.ID)
		exchangeCntAttr.SetInt( key, exchangeCntAttr.GetInt(key) + 1 )
	}
}

func (c *activityCom) forEachExchangeCnt(aid int, callback func(goodsID, cnt int)) {
	exchangeCntAttr := c.getExchangeCntAttr(aid)
	if exchangeCntAttr == nil {
		return
	}
	exchangeCntAttr.ForEachKey(func(key string) {
		goodsID, _ := strconv.Atoi(key)
		callback(goodsID, exchangeCntAttr.GetInt(key))
	})
}

func (c *activityCom) getActivityVersion(aid int) int {
	attr := c.getActivityAttr(aid)
	if attr == nil {
		return 0
	}
	return attr.GetInt(aTypes.Version)
}

func (c *activityCom) setActivityVersion(aid int) {
	activity := mod.IAMod.GetActivityByID(aid)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("spring activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	c.getActivityAttr(aid).SetInt(aTypes.Version, activity.GetActivityVersion())
}

func (c *activityCom) checkVersion(aid int) bool {
	act := aTypes.IMod.GetActivityByID(aid)
	if act == nil {
		glog.Errorf("spring activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
		return false
	}
	if act.GetActivityVersion() == c.getActivityVersion(aid) {
		return true
	}
	c.attr.Del(strconv.Itoa(aid))
	c.setActivityVersion(aid)
	return false
}
