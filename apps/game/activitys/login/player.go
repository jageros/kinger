package login

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	//"kinger/gopuppy/common/timer"
	aTypes "kinger/apps/game/activitys/types"
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
		attr:   pcm.InitAttr(aTypes.LoginActivity),
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
		glog.Errorf("Login activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	attr.SetInt(aTypes.Version, activity.GetActivityVersion())
}

func (c *activityCom) getTotalLoginDay(aid int) int {
	num := 1
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
		num = 1
	}
	if !c.checkVersion(aid) {
		num = 1
	}
	totalDay := attr.GetInt(aTypes.TotalLoginDay)
	if totalDay > 0 {
		num = totalDay
	}
	if num == 1 {
		attr.SetInt(aTypes.TotalLoginDay, num)
	}
	return num
}

func (c *activityCom) onCrossDayUpdateLoginData() {
	idList := mod.IAMod.GetActivityIdList()
	for _, id := range idList {
		aid := strconv.Itoa(id)
		c.checkVersion(id)
		if c.ipc.Conform(id) {
			attr := c.attr.GetMapAttr(aid)
			if attr == nil {
				attr = attribute.NewMapAttr()
				c.attr.SetMapAttr(aid, attr)
				attr.SetInt(aTypes.TotalLoginDay, 1)
				//attr.SetInt(aTypes.ContinueLoginDay, 1)
			} else {
				oldNum := attr.GetInt(aTypes.TotalLoginDay)
				newNum := oldNum + 1
				attr.SetInt(aTypes.TotalLoginDay, newNum)
				c.ipc.PushFinshNum(id, 0, newNum)
				c.updateReceiveStatus(id, oldNum, newNum)
			}
		}
	}
}

func (c *activityCom) hasReceive(activityID, rewardID int) (bool, error) {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Login activity hasReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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

	rwAttr := attr.GetMapAttr(aTypes.LoginRewardBool)
	if rwAttr == nil {
		return false, nil
	}

	return rwAttr.GetBool(rid), nil
}

func (c *activityCom) setReceive(activityID, rewardID int) error {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Login activity setReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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

	rwAttr := attr.GetMapAttr(aTypes.LoginRewardBool)
	if rwAttr == nil {
		rwAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.LoginRewardBool, rwAttr)
	}
	rwAttr.SetBool(rid, true)
	return nil
}

func (c *activityCom) updateHint() {
	for id, rw := range mod.id2Reward {
		potNum := 0
		if !c.ipc.Conform(id) {
			continue
		}
		for k, _ := range rw.ActivityLoginRewardMap {
			if c.conformRewardCondition(id, k) {
				potNum++
			}
		}
		c.player.UpdateHint(pb.HintType_HtActivity+pb.HintType(id), potNum)
	}
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

func (c *activityCom) conformRewardCondition(activityID, rewardID int) bool {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("Login activity conformRewardCondition GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	received, err := c.hasReceive(activityID, rewardID)
	if err != nil {
		glog.Errorf("Login activity conformRewardCondition hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	if !c.ipc.Conform(activityID) || received {
		return false
	}

	needLoginDay, err := mod.getRewardFinshCondition(activityID, rewardID)
	if err != nil {
		glog.Errorf("Login activity conformRewardCondition getRewardFinshCondition err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	totalLoginDay := c.getTotalLoginDay(activityID)

	if totalLoginDay >= needLoginDay {
		c.ipc.LogActivity(activityID, activity.GetActivityVersion(), activity.GetActivityType(), rewardID, aTypes.ActivityOnFinsh)
		return true
	}
	return false
}

func (c *activityCom) getRewardReceiveStatus(aid, rid int) (rst pb.ActivityReceiveStatus) {
	hasReceive, err := c.hasReceive(aid, rid)
	if err != nil {
		glog.Errorf("Login activity getRewardReceiveStatus hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
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

func (c *activityCom) checkVersion(aid int) bool {
	act := mod.IAMod.GetActivityByID(aid)
	if act == nil {
		glog.Errorf("login activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
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
			glog.Error("Login activity updateReceiveStatus getRewardFinshCondition err=%s, uid=%d, activityID=%d, RewardID=%d", err, c.player.GetUid(), aid, rid)
		}
		if fnum > oldNum && fnum <= newNum {
			rst := c.getRewardReceiveStatus(aid, rid)
			c.ipc.PushReceiveStatus(aid, rid, rst)
		}
	}
}


func (c *activityCom) isClosed (aid int) bool {
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

func (c *activityCom) setClosed (aid int) {
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
		if len(rw.ActivityLoginRewardMap) <= 0 {
			continue
		}
		for rid, _ := range rw.ActivityLoginRewardMap {
			b, err := c.hasReceive(aid, rid)
			if err != nil {
				glog.Errorf("Login activity updateActivityCloseStatus hasReceive return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
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
	c.onCrossDayUpdateLoginData()
	c.updateHint()
}
