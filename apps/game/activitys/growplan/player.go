package growplan

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
		attr:   pcm.InitAttr(aTypes.GrowPlanActivity),
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
		glog.Errorf("growplan activity setActivityVersion GetActivityByID err=%s, uid=%d, activityID=%d", err, c.player.GetUid(), aid)
		return
	}
	attr.SetInt(aTypes.Version, activity.GetActivityVersion())
}

func (c *activityCom) setBuyGift(aid int, goodsId string) {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	gAttr := attr.GetMapAttr(aTypes.GrowPlan_hasBuyGift)
	if gAttr == nil {
		gAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.GrowPlan_hasBuyGift, gAttr)
	}
	gAttr.SetBool(goodsId, true)
}

func (c *activityCom) isBuyGift(aid, rid int) bool {
	if !c.checkVersion(aid) {
		return false
	}
	goodsId := mod.getRewardData(aid, rid).Purchase
	if goodsId == "" {
		return true
	}
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return false
	}
	gAttr := attr.GetMapAttr(aTypes.GrowPlan_hasBuyGift)
	if gAttr == nil {
		return false
	}
	return gAttr.GetBool(goodsId)
}

func (c *activityCom) get2ArgAttrNum(aid int, key string, arg int) int {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return 0
	}
	ccAttr := attr.GetMapAttr(key)
	if ccAttr == nil {
		return 0
	}
	argStr := strconv.Itoa(arg)
	return ccAttr.GetInt(argStr)
}

func (c *activityCom) set2ArgAttrNum(aid int, key string, key2 int, num int) {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	key2Str := strconv.Itoa(key2)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	ccAttr := attr.GetMapAttr(key)
	if ccAttr == nil {
		ccAttr = attribute.NewMapAttr()
		attr.SetMapAttr(key, ccAttr)
	}
	ccAttr.SetInt(key2Str, num)
}

func (c *activityCom) get1ArgAttrNum(aid int, key string) int {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		return 0
	}
	return attr.GetInt(key)
}

func (c *activityCom) set1ArgAttrNum(aid int, key string, num int) {
	c.checkVersion(aid)
	aidStr := strconv.Itoa(aid)
	attr := c.attr.GetMapAttr(aidStr)
	if attr == nil {
		attr = attribute.NewMapAttr()
		c.attr.SetMapAttr(aidStr, attr)
	}
	attr.SetInt(key, num)
}

func (c *activityCom) setContinuousWinNum(aid int, isWin bool) {
	rids := mod.getRewardIdList(aid)
	lvl := c.player.GetPvpLevel()
	for _, rid := range rids {
		ty, val, _ := mod.getRewardFinshCondition(aid, rid)
		if ty != ty_ContinuousWinNum_ {
			continue
		}
		if lvl < val[0] {
			continue
		}
		num := c.get2ArgAttrNum(aid, aTypes.GrowPlan_continuousWinNum, val[0])
		if num >= val[1] {
			continue
		}
		if isWin {
			c.set2ArgAttrNum(aid, aTypes.GrowPlan_continuousWinNum, val[0], num+1)
		} else {
			c.set2ArgAttrNum(aid, aTypes.GrowPlan_continuousWinNum, val[0], 0)
		}

	}
}

func (c *activityCom) getFinshNum(aid, rid int) int {
	ty, val, err := mod.getRewardFinshCondition(aid, rid)
	if err != nil {
		glog.Errorf("GrowPlan activity getFinshNum getRewardFinshCondition err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
		return 0
	}
	switch ty {
	case ty_useCampCardWinBattle_:
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_campCardWin, val[0])
	case ty_achieveLevel_:
		return c.player.GetMaxPvpLevel()
	case ty_hasLevelCard_:
		return c.player.CalcCollectCardNumByLevel(val[1])
	case ty_hasStarCard_:
		return c.player.CalcCollectCardNumByStar(val[1])
	case ty_hasFriends_:
		return c.player.GetFriendsNum()
	case ty_openTreasure_:
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_TreasureOpenNum)
	case ty_jadeConsume_:
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_jadeConsume)
	case ty_combat_:
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_combat)
	case ty_watchBattleReport_:
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_watchBattleReportNum)
	case ty_sendBattleReport_:
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_sendBattleReportNum)
	case ty_hitOutStarCard_:
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_hitOutStarCardNum, val[1])
	case ty_hitOutCampCard_:
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_hitOutCampCardNum, val[1])
	case ty_useCampWin_:
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_useCampWinNum, val[0])
	case ty_useCampBattle_:
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_useCampBattleNum, val[0])
	case ty_finshMission_:
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_finshMissionNum)
	case ty_totalRecharge_: //16 //累计充值{0}元({1}/{2}
		return c.get1ArgAttrNum(aid, aTypes.GrowPlan_totalRecharge)
	case ty_totalBuyJade_: //17 //累计购买{0}宝玉({1}/{2})

	case ty_vipExclusive_: // 18 //月卡专属奖励
		if c.player.IsVip() {
			return 1
		}
		return 0
	case ty_login_: //  19 //{0}期间登陆游戏

	case ty_battleNum_: // 20 //天梯对战{0}次
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_useCampBattleNum, 0)
	case ty_battlewinNum: //21 //天梯对战胜利{0}次
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_useCampWinNum, 0)
	case ty_BuyLimtGift_: //22 //购买后领取
		if c.isBuyGift(aid, rid) {
			return 1
		}
		return 0
	case ty_ContinuousWinNum_: //23 //在{0}连续胜利{1}场
		return c.get2ArgAttrNum(aid, aTypes.GrowPlan_continuousWinNum, val[0])
	}
	return 0
}

func (c *activityCom) setReceive(activityID, rewardID int) error {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("growplan activity setReceive GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
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

	rwAttr := attr.GetMapAttr(aTypes.GrowPlan_rewardRechiceBool)
	if rwAttr == nil {
		rwAttr = attribute.NewMapAttr()
		attr.SetMapAttr(aTypes.GrowPlan_rewardRechiceBool, rwAttr)
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

	rwAttr := attr.GetMapAttr(aTypes.GrowPlan_rewardRechiceBool)
	if rwAttr == nil {
		return false, nil
	}

	return rwAttr.GetBool(rid), nil
}

func (c *activityCom) conformRewardCondition(activityID, rewardID int) bool {
	activity := mod.IAMod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("GrowPlan activity conformRewardCondition GetActivityByID err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}

	received, err := c.hasReceive(activityID, rewardID)
	if err != nil {
		glog.Errorf("GrowPlan activity conformRewardCondition1 hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	lastReceived := true
	if rewardID > 1 {
		lastReceived, err = c.hasReceive(activityID, rewardID-1)
		if err != nil {
			glog.Errorf("GrowPlan activity conformRewardCondition2 hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
			return false
		}
	}

	if !c.ipc.Conform(activityID) || received || !lastReceived {
		return false
	}

	_, needNum, err := mod.getRewardFinshCondition(activityID, rewardID)
	if err != nil {
		glog.Errorf("GrowPlan activity conformRewardCondition getRewardFinshCondition err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), activityID, rewardID)
		return false
	}
	hasNum := c.getFinshNum(activityID, rewardID)
	if hasNum >= needNum[1] {
		return true
	}
	return false
}

func (c *activityCom) getRewardReceiveStatus(aid, rid int) (rst pb.ActivityReceiveStatus) {
	hasReceive, err := c.hasReceive(aid, rid)
	if err != nil {
		glog.Errorf("GrowPlan activity getRewardReceiveStatus hasReceive err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
		return
	}
	if hasReceive {
		return pb.ActivityReceiveStatus_HasReceive
	}
	canReceive := c.conformRewardCondition(aid, rid)
	if canReceive {
		if c.isBuyGift(aid, rid) {
			activity := mod.IAMod.GetActivityByID(aid)
			c.ipc.LogActivity(aid, activity.GetActivityVersion(), activity.GetActivityType(), rid, aTypes.ActivityOnFinsh)
			return pb.ActivityReceiveStatus_CanReceive
		}
		return pb.ActivityReceiveStatus_NeedRecharge
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
			return
		}
		for k, _ := range rw.ID2ActivityGrowPlanReward {
			receStatus := c.getRewardReceiveStatus(id, k)
			if receStatus == pb.ActivityReceiveStatus_CanReceive {
				potNum++
			}
			c.ipc.PushReceiveStatus(id, k, receStatus)
		}
		c.player.UpdateHint(pb.HintType_HtActivity+pb.HintType(id), potNum)
	}
}

func (c *activityCom) checkVersion(aid int) bool {
	act := mod.IAMod.GetActivityByID(aid)
	if act == nil {
		glog.Errorf("growplan activity get activity by id error, activityID=%d, uid=%d", aid, c.player.GetUid())
		return false
	}
	if act.GetActivityVersion() == c.getActivityVersion(aid) {
		return true
	}
	c.attr.Del(strconv.Itoa(aid))
	c.setActivityVersion(aid)
	return false
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
		if len(rw.ID2ActivityGrowPlanReward) <= 0 {
			continue
		}
		for rid, _ := range rw.ID2ActivityGrowPlanReward {
			b, err := c.hasReceive(aid, rid)
			if err != nil {
				glog.Errorf("growplan activity updateActivityCloseStatus hasReceive return err=%s, uid=%d, activityID=%d, rewardID=%d", err, c.player.GetUid(), aid, rid)
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

func (c *activityCom) pushFinshNum(aid, ty int) {
	rids := mod.getRewardIdList(aid)
	for _, rid := range rids {
		types, _, _ := mod.getRewardFinshCondition(aid, rid)
		if ty == ty_unknowType_ || ty == types {
			num := c.getFinshNum(aid, rid)
			c.ipc.PushFinshNum(aid, rid, num)
		}

	}
}

func (c *activityCom) OnCrossDays(dayno int) {
	if dayno == c.player.GetDataDayNo() {
		return
	}
	c.updateActivityCloseStatus()
}
