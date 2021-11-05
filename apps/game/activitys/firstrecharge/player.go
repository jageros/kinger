package firstrecharge

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"kinger/proto/pb"
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
		attr:   pcm.InitAttr(aTypes.FirstRechargeActivity),
	}
}

func (c *activityCom) getActivityVersion(activityID int) int {
	aid := strconv.Itoa(activityID)
	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		return 0
	}
	return attr.GetInt(aTypes.Version)
}

func (c *activityCom) setActivityVersion(activityID int) {
	aid := strconv.Itoa(activityID)
	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aid, attr)
	}
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("First Recharge activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	attr.SetInt(aTypes.Version, activity.GetActivityVersion())
}


func (c *activityCom) setReceive(activityID, rewardID int) error {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Recharge activity setReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return err
	}
	aid := strconv.Itoa(activityID)
	rid := strconv.Itoa(rewardID)

	c.checkVersion(activityID)

	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aid, attr)
	}

	rwAttr := attr.GetMapAttr(aTypes.FirstRechargeRewardBool)
	if rwAttr == nil {
		rwAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.FirstRechargeRewardBool, rwAttr)
	}
	rwAttr.SetBool(rid, true)
	return nil
}

func (c *activityCom) hasReceive(activityID, rewardID int) (bool, error) {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		return false, err
	}
	aid := strconv.Itoa(activityID)
	rid := strconv.Itoa(rewardID)

	if !c.checkVersion(activityID) {
		return false, nil
	}
	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		return false, nil
	}

	rwAttr := attr.GetMapAttr(aTypes.FirstRechargeRewardBool)
	if rwAttr == nil {
		return false, nil
	}

	return rwAttr.GetBool(rid), nil
}

func (c *activityCom) conformRewardCondition(activityID, rewardID int) bool {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("First Recharge activity conformRewardCondition GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	received, err := c.hasReceive(activityID, rewardID)
	if err != nil {
		glog.Errorf("First Recharge activity conformRewardCondition hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	if !c.ipc.ConformOpen(activityID) || !c.ipc.ConformTime(activityID) || received {
		return false
	}

	isRecharge := c.hasBuy(activityID, rewardID)

	if isRecharge {
		c.ipc.LogActivity(activityID, activity.GetActivityVersion(), activity.GetActivityType(), rewardID, aTypes.ActivityOnFinsh)
		return true
	}
	return false
}

func (c *activityCom) getRewardReceiveStatus(aid, rid int) (rst pb.ActivityReceiveStatus) {
	hasReceive, err := c.hasReceive(aid, rid)
	if err != nil {
		glog.Errorf("First Recharge activity getRewardReceiveStatus hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
		return
	}
	if hasReceive {
		return pb.ActivityReceiveStatus_HasReceive
	}
	canReceive := c.conformRewardCondition(aid, rid)
	if canReceive {
		return pb.ActivityReceiveStatus_CanReceive
	}
	return pb.ActivityReceiveStatus_CanNotReceive
}

func (c *activityCom) giveReward(activityID, rewardID int, rd *pb.Reward) error {
	rw := mod.getRewardMap(activityID, rewardID)
	for stuff, num := range rw {
		err := c.ipc.GiveReward(stuff, int(num), rd, activityID, rewardID)
		if err != nil {
			return err
		}
	}
	c.setReceive(activityID, rewardID)
	potNum := c.player.GetHintCount(pb.HintType_HtFirstRecharge)
	c.player.UpdateHint(pb.HintType_HtFirstRecharge, potNum-1)
	return nil
}

func (c *activityCom) updateHint() {
	for id, rw := range mod.id2Reward {
		if !c.ipc.ConformTime(id) || !c.ipc.ConformOpen(id) {
			continue
		}
		potNum := 0
		for k, _ := range rw.ID2ActivityFirstRechargeReward {
			if rst:= c.getRewardReceiveStatus(id, k); rst == pb.ActivityReceiveStatus_CanReceive {
				potNum++
			}
		}
		c.player.UpdateHint(pb.HintType_HtFirstRecharge, potNum)
	}
}

func (c *activityCom) updateActivityTagList() {
	tagList := c.ipc.GetActivityTagList()
	for aid, _ := range mod.id2Reward {
		for k, v := range tagList {
			if v == int32(aid) {
				tagList = append(tagList[:k], tagList[k+1:]...)
			}
		}
	}
	c.ipc.UpdateActivityTagList(tagList)
}

func (c *activityCom) checkVersion(aid int) bool {
	act := mod.IAMod.GetActivityByID(aid)
	if act == nil {
		glog.Errorf("First recharge activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
		return false
	}
	if act.GetActivityVersion() == c.getActivityVersion(aid) {
		return true
	}
	c.attr.Del(strconv.Itoa(aid))
	c.setActivityVersion(aid)
	return false
}

func (c *activityCom) conform(aid int) bool {
	return c.ipc.ConformTime(aid) && c.ipc.ConformOpen(aid)
}

func (c *activityCom) getRechargePrice(resID string) int {
	return module.Shop.GetLimitGiftPrice(resID, c.player)
}

func (c *activityCom) setBuyGoods(aid int, resId string) {
	aidStr := strconv.Itoa(aid)
	c.checkVersion(aid)
	if c.conform(aid) {
		attr := c.attr.GetMapAttr(aidStr)
		if attr == nil {
			attr = attribute.NewMapAttr()
			rAttr := attribute.NewMapAttr()
			c.attr.SetMapAttr(aidStr, attr)
			attr.SetMapAttr(aTypes.FirstRechargeHasGoodsBool, rAttr)
			rAttr.SetBool(resId, true)
		} else {
			rAttr := attr.GetMapAttr(aTypes.FirstRechargeHasGoodsBool)
			if rAttr == nil {
				rAttr = attribute.NewMapAttr()
				attr.SetMapAttr(aTypes.FirstRechargeHasGoodsBool, rAttr)
				rAttr.SetBool(resId, true)
			} else {
				rAttr.SetBool(resId, true)
			}
		}
	}

}

func (c *activityCom) hasBuy(aid, rid int) bool {
	if !c.checkVersion(aid) {
		return false
	}
	rd := mod.getRewardData(aid, rid)
	if rd == nil {
		err := gamedata.GameError(aTypes.GetRewardError)
		glog.Errorf("FirstRecharge getLoginRechargeBool getRewardData err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
		return false
	}
	resID := rd.GoodsID
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return false
	}
	rAttr := attr.GetMapAttr(aTypes.FirstRechargeHasGoodsBool)
	if rAttr == nil {
		return false
	}
	return rAttr.GetBool(resID)
}