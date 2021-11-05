package dailyshare

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
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
		attr:   pcm.InitAttr(aTypes.DailyShareActivity),
	}
}

func (c *activityCom) getActivityVersion(aid int) int {
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return 0
	}
	return attr.GetInt(aTypes.Version)
}

func (c *activityCom) setActivityVersion(aid int) {
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	activity := mod.IAMod.GetActivityByID(aid)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("DailyRecharge activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	attr.SetInt(aTypes.Version, activity.GetActivityVersion())
}

func (c *activityCom) getTodayShareAmount(aid int) int {
	if !c.checkVersion(aid) {
		return 0
	}
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return 0
	}
	return attr.GetInt(aTypes.TodayShareAmount)
}

func (c *activityCom) setTodayShareAmount(activityID, num int) {
	c.checkVersion(activityID)
	aid := strconv.Itoa(activityID)
	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aid, attr)
	}
	oldNum := attr.GetInt(aTypes.TodayShareAmount)
	newNum := oldNum + num
	attr.SetInt(aTypes.TodayShareAmount, newNum)
	c.ipc.PushFinshNum(activityID, 0, newNum)
	c.updateReceiveStatus(activityID, oldNum, newNum)
}

func (c *activityCom) setReceive(activityID, rewardID int) error {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("DailyRecharge activity setReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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

	rwAttr := attr.GetMapAttr(aTypes.DailyShareReceiveBool)
	if rwAttr == nil {
		rwAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.DailyShareReceiveBool, rwAttr)
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

	rwAttr := attr.GetMapAttr(aTypes.DailyShareReceiveBool)
	if rwAttr == nil {
		return false, nil
	}

	return rwAttr.GetBool(rid), nil
}

func (c *activityCom) conformRewardCondition(activityID, rewardID int) bool {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("DailyShare activity conformRewardCondition GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	received, err := c.hasReceive(activityID, rewardID)
	if err != nil {
		glog.Errorf("DailyShare activity conformRewardCondition hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	if !c.ipc.Conform(activityID) || received {
		return false
	}

	needRecharge, err := mod.getRewardFinshCondition(activityID, rewardID)
	if err != nil {
		glog.Errorf("DailyShare activity conformRewardCondition getRewardFinshCondition err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	rechargeAmount := c.getTodayShareAmount(activityID)

	if rechargeAmount >= needRecharge {
		c.ipc.LogActivity(activityID, activity.GetActivityVersion(), activity.GetActivityType(), rewardID, aTypes.ActivityOnFinsh)
		return true
	}
	return false
}

func (c *activityCom) getRewardReceiveStatus(aid, rid int) (rst pb.ActivityReceiveStatus) {
	hasReceive, err := c.hasReceive(aid, rid)
	if err != nil {
		glog.Errorf("Recharge activity getRewardReceiveStatus hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
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
	potNum := c.player.GetHintCount(pb.HintType_HtActivity + pb.HintType(activityID))
	c.player.UpdateHint(pb.HintType_HtActivity+pb.HintType(activityID), potNum-1)
	return nil
}

func (c *activityCom) updateHint() {
	for id, rw := range mod.id2Reward {
		potNum := 0
		if !c.ipc.Conform(id) {
			continue
		}
		for k, _ := range rw.ID2ActivityDailyShareReward {
			if c.conformRewardCondition(id, k) {
				potNum++
			}
		}
		c.player.UpdateHint(pb.HintType_HtActivity+pb.HintType(id), potNum)
	}
}

func (c *activityCom) checkVersion(aid int) bool {
	act := mod.IAMod.GetActivityByID(aid)
	if act == nil {
		glog.Errorf("DailyShare activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
		return false
	}
	if act.GetActivityVersion() == c.getActivityVersion(aid) {
		return true
	}
	c.attr.Del(strconv.Itoa(aid))
	c.setActivityVersion(aid)
	return false
}

func (c *activityCom) updateReceiveStatus(aid, oldNum, newNum int) {
	rids := mod.getRewardIdList(aid)
	for _, rid := range rids {
		fnum, err := mod.getRewardFinshCondition(aid, rid)
		if err != nil {
			glog.Error("DailyShare activity updateReceiveStatus getRewardFinshCondition err=%s, uid=%d, activityID=%d, RewardID=%d", err, c.player.GetUid(), aid, rid)
		}
		if fnum > oldNum && fnum <= newNum {
			rst := c.getRewardReceiveStatus(aid, rid)
			c.ipc.PushReceiveStatus(aid, rid, rst)
		}
	}
}


func (c *activityCom) OnCrossDays(dayno int) {
	if dayno == c.player.GetDataDayNo() {
		return
	}
	aTypes.IMod.ForEachActivityDataByType(consts.ActivityOfDailyShare, func(data aTypes.IActivity) {
		c.attr.Del(strconv.Itoa(data.GetActivityId()))
	})
}