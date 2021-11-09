package consume

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
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
		attr:   pcm.InitAttr(aTypes.ConsumeActivity),
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
		glog.Errorf("Consume activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	attr.SetInt(aTypes.Version, activity.GetActivityVersion())
}

func (c *activityCom) getTotalConsumeAmount(activityID int) int {
	if !c.checkVersion(activityID) {
		return 0
	}
	aid := strconv.Itoa(activityID)
	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		return 0
	}
	return attr.GetInt(aTypes.TotalConsumeAmount)
}

func (c *activityCom) setTotalConsumeAmount(activityID, money int) {
	c.checkVersion(activityID)
	aid := strconv.Itoa(activityID)
	attr := c.attr.GetMapAttr(aid)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aid, attr)
	}
	oldm := attr.GetInt(aTypes.TotalConsumeAmount)
	newm := oldm + money
	attr.SetInt(aTypes.TotalConsumeAmount, newm)
	c.ipc.PushFinshNum(activityID, 0, newm)
	c.updateReceiveStatus(activityID, oldm, newm)
}

func (c *activityCom) setReceive(activityID, rewardID int) error {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Consume activity setReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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

	rwAttr := attr.GetMapAttr(aTypes.ConsumeRewardBool)
	if rwAttr == nil {
		rwAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.ConsumeRewardBool, rwAttr)
	}
	rwAttr.SetBool(rid, true)
	return nil
}

func (c *activityCom) hasReceive(activityID, rewardID int) (bool, error) {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Consume activity hasReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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

	rwAttr := attr.GetMapAttr(aTypes.ConsumeRewardBool)
	if rwAttr == nil {
		return false, nil
	}

	return rwAttr.GetBool(rid), nil
}

func (c *activityCom) conformRewardCondition(activityID, rewardID int) bool {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Consume activity conformRewardCondition GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	received, err := c.hasReceive(activityID, rewardID)
	if err != nil {
		glog.Errorf("Consume activity conformRewardCondition hasReceive return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	if !c.ipc.Conform(activityID) || received {
		return false
	}

	needNum, err := mod.getRewardFinshCondition(activityID, rewardID)
	if err != nil {
		glog.Errorf("Consume activity conformRewardCondition getRewardFinshCondition return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	hasNum := c.getTotalConsumeAmount(activityID)

	if hasNum >= needNum {
		c.ipc.LogActivity(activityID, activity.GetActivityVersion(), activity.GetActivityType(), rewardID, aTypes.ActivityOnFinsh)
		return true
	}
	return false
}

func (c *activityCom) getRewardReceiveStatus(aid, rid int) (rst pb.ActivityReceiveStatus) {
	hasReceive, err := c.hasReceive(aid, rid)
	if err != nil {
		glog.Errorf("Consume activity getRewardReceiveStatus hasReceive return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
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
			glog.Errorf("Consume activity giveReward GiveReward return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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
		for k, _ := range rw.ActivityConsumeRewardMap {
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
		glog.Errorf("consume activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
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
			glog.Error("Consume activity updateReceiveStatus getRewardFinshCondition err=%s, uid=%d, activityID=%d, RewardID=%d", err, c.player.GetUid(), aid, rid)
		}
		if fnum > oldNum && fnum <= newNum {
			rst := c.getRewardReceiveStatus(aid, rid)
			c.ipc.PushReceiveStatus(aid, rid, rst)
		}
	}
}

func (c *activityCom) isClosed(aid int) bool {
	if !c.checkVersion(aid) {
		return false
	}
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return false
	}
	return attr.GetBool(aTypes.CloseStatus)
}

func (c *activityCom) setClosed(aid int) {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	attr.SetBool(aTypes.CloseStatus, true)
}

func (c *activityCom) updateActivityTagList() {
	tagList := c.ipc.GetActivityTagList()
	for k, aid := range tagList {
		act := mod.IAMod.GetActivityByID(int(aid))
		if act != nil && c.isClosed(int(aid)) {
			tagList = append(tagList[:k], tagList[k+1:]...)
		}
	}
	c.ipc.UpdateActivityTagList(tagList)
}

func (c *activityCom) updateActivityCloseStatus() {
	for aid, rw := range mod.id2Reward {
		flag := 0
		if len(rw.ActivityConsumeRewardMap) <= 0 {
			continue
		}
		for rid, _ := range rw.ActivityConsumeRewardMap {
			b, err := c.hasReceive(aid, rid)
			if err != nil {
				glog.Errorf("Consume activity updateActivityCloseStatus hasReceive return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
				break
			}
			if !b {
				flag = 1
				break
			}
		}
		if flag == 1 {
			continue
		}
		c.setClosed(aid)
	}
}

func (c *activityCom) OnCrossDays(dayno int) {
	if dayno == c.player.GetDataDayNo() {
		return
	}
	c.updateActivityCloseStatus()
}
