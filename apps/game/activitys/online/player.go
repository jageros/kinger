package online

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"time"
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
		attr:   pcm.InitAttr(aTypes.OnlineActivity),
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
		glog.Errorf("Online activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	attr.SetInt(aTypes.Version, activity.GetActivityVersion())
}

func (c *activityCom) getLastReceiveTime(aid, rid int) int64 {
	if !c.checkVersion(aid) {
		return 0
	}
	aidStr := strconv.Itoa(aid)
	ridStr := strconv.Itoa(rid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return 0
	}
	rwAttr := attr.GetMapAttr(aTypes.LastReceiveTime)
	if rwAttr == nil {
		return 0
	}
	return rwAttr.GetInt64(ridStr)
}

func (c *activityCom) setLastReceiveTime(aid, rid int) {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	ridStr := strconv.Itoa(rid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	rwAttr := attr.GetMapAttr(aTypes.LastReceiveTime)
	if rwAttr == nil {
		rwAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.LastReceiveTime, rwAttr)
	}
	rwAttr.SetInt64(ridStr, time.Now().Unix())
}

func (c *activityCom) hasReceive(aid, rid int) (bool, error) {
	activity := mod.IAMod.GetActivityByID(aid)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Online activity hasReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
		return false, err
	}

	if !c.checkVersion(aid) {
		return false, nil
	}
	lastTim := c.getLastReceiveTime(aid, rid)
	b := utils.IsSameDay(time.Now().Unix(), lastTim)
	return b, nil
}

func (c *activityCom) conformRewardCondition(activityID, rewardID int) bool {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Online activity conformRewardCondition GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	received, err := c.hasReceive(activityID, rewardID)
	if err != nil {
		glog.Errorf("Online activity conformRewardCondition hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	if !c.ipc.Conform(activityID) || received {
		return false
	}

	needTim, err := mod.getRewardFinshCondition(activityID, rewardID)
	if err != nil {
		glog.Errorf("Online activity conformRewardCondition getRewardFinshCondition err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	startTime, endTime, err := changeStartEndHour2Time(needTim)
	if err != nil {
		glog.Errorf("Online activity conformRewardCondition changeStartEndHour2Time err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	if time.Now().After(startTime) && time.Now().Before(endTime) {
		c.ipc.LogActivity(activityID, activity.GetActivityVersion(), activity.GetActivityType(), rewardID, aTypes.ActivityOnFinsh)
		return true
	}
	return false
}

func (c *activityCom) getRewardReceiveStatus(aid, rid int) (rst pb.ActivityReceiveStatus) {
	hasReceive, err := c.hasReceive(aid, rid)
	if err != nil {
		glog.Errorf("Online activity getRewardReceiveStatus hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
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

func (c *activityCom) giveReward(aid, rid int, rd *pb.Reward) error {
	rw := mod.getRewardMap(aid, rid)
	for stuff, num := range rw {
		if aid == 4 && stuff == "ticket" {
			num = int32(module.OutStatus.BuffAccTreasureCntByActivity(c.player, int(num)))
		}
		err := c.ipc.GiveReward(stuff, int(num), rd, aid, rid)
		if err != nil {
			return err
		}
	}
	c.setLastReceiveTime(aid, rid)
	potNum := c.player.GetHintCount(pb.HintType_HtActivity + pb.HintType(aid))
	c.player.UpdateHint(pb.HintType_HtActivity+pb.HintType(aid), potNum-1)
	return nil
}

func (c *activityCom) updateHint() {
	for aid, rw := range mod.id2Reward {
		potNum := 0
		if !c.ipc.Conform(aid) {
			continue
		}
		for rid, _ := range rw.ActivityOnlineRewardMap {
			if c.conformRewardCondition(aid, rid) {
				potNum++
			}
		}
		c.player.UpdateHint(pb.HintType_HtActivity+pb.HintType(aid), potNum)
	}
}

func (c *activityCom) checkVersion(aid int) bool {
	act := mod.IAMod.GetActivityByID(aid)
	if act == nil {
		glog.Errorf("online activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
		return false
	}
	if act.GetActivityVersion() == c.getActivityVersion(aid) {
		return true
	}
	c.attr.Del(strconv.Itoa(aid))
	c.setActivityVersion(aid)
	return false
}

func (c *activityCom) updateReceiveStatus() {
	aids := mod.IAMod.GetActivityIdList()
	for _, aid := range aids {
		if aTypes.IsInArry(int32(aid), c.ipc.GetActivityTagList()) {
			rids := mod.getRewardIdList(aid)
			for _, rid := range rids {
				rst := c.getRewardReceiveStatus(aid, rid)
				c.ipc.PushReceiveStatus(aid, rid, rst)
			}
		}
	}
}
