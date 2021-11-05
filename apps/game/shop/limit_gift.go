package shop

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"time"
	"kinger/gopuppy/common/evq"
)

const (
	minVipCardGiftID = "minivip"
	timeLayout = "2006-01-02 15:04:05"
)

type iLimitGift interface {
	buy() (reply proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType)
	getGameData() *gamedata.LimitGift
	tryRefresh(now time.Time)
	packMsg() *pb.LimitGift
	getGiftID() string
	setIsNew(val bool)
	isBuy() bool
	canShow() bool
	isNew() bool
	onReborn() bool
	reset()
	canBuy() bool
	onCrossDay()
	onLogin()
	setVersion(version int)
	addShowCondition(condition limitGiftCondition)
	addHideCondition(condition limitGiftCondition)
	addBuyCondition(condition limitGiftCondition)
}

type limitGiftCondition func(player types.IPlayer) bool

type limitGiftMgrSt struct {
	player types.IPlayer
	gifts []iLimitGift
	prefix2TeamGift map[string]*teamGiftSt
	limitGiftsAttr *attribute.ListAttr
}

func newLimitGiftMgr(player types.IPlayer, cptAttr *attribute.MapAttr) *limitGiftMgrSt {
	lgm := &limitGiftMgrSt{
		player: player,
		prefix2TeamGift: map[string]*teamGiftSt{},
	}

	lgm.limitGiftsAttr = cptAttr.GetListAttr("limitGifts")
	if lgm.limitGiftsAttr == nil {
		lgm.limitGiftsAttr = attribute.NewListAttr()
		cptAttr.SetListAttr("limitGifts", lgm.limitGiftsAttr)
	} else {
		lgm.limitGiftsAttr.ForEachIndex(func(index int) bool {
			giftAttr := lgm.limitGiftsAttr.GetMapAttr(index)
			gift, _ := lgm.newLimitGiftByAttr(giftAttr, player)
			if gift != nil {
				lgm.gifts = append(lgm.gifts, gift)
			}
			return true
		})
	}

	return lgm
}

func (lgm *limitGiftMgrSt) newCondition(conditionInfo []string) limitGiftCondition {
	if len(conditionInfo) < 2 {
		return nil
	}

	switch conditionInfo[0] {
	case "team":
		team, _ := strconv.Atoi(conditionInfo[1])
		return func(player types.IPlayer) bool {
			return player.GetPvpTeam() >= team
		}

	case "maxTeam":
		team, _ := strconv.Atoi(conditionInfo[1])
		return func(player types.IPlayer) bool {
			rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
			return rankGameData.Ranks[player.GetMaxPvpLevel()].Team >= team
		}

	case "bought":
		return func(player types.IPlayer) bool {
			if player.GetPvpLevel() < 2 {
				return false
			}
			for _, gift := range lgm.gifts {
				data := gift.getGameData()
				if data == nil || data.GiftIDPrefix != conditionInfo[1] {
					continue
				}

				if gift.canShow() && gift.canBuy() {
					return false
				}
			}
			return true
		}

	default:
		return nil
	}
}

func (lgm *limitGiftMgrSt) packMsg() []*pb.LimitGift {
	var gifts []*pb.LimitGift
	lgm.forEachCanShopShowLimitGifts(func(gift iLimitGift) {
		msg := gift.packMsg()
		if msg != nil {
			gifts = append(gifts, msg)
		}
	})
	return gifts
}

func (lgm *limitGiftMgrSt) forEachCanShopShowLimitGifts(callback func(gift iLimitGift)) {
	for _, gift := range lgm.gifts {
		if !gift.isBuy() && gift.canShow() && gift.getGameData().Visible {
			callback(gift)
		}
	}
}

func (lgm *limitGiftMgrSt) forEachCanShowLimitGifts(callback func(gift iLimitGift)) {
	for _, gift := range lgm.gifts {
		if gift.canShow() {
			callback(gift)
		}
	}
}

func (lgm *limitGiftMgrSt) getLimitGiftByTreasure(treasureModelID string) iLimitGift {
	for _, gift := range lgm.gifts {
		giftData := gift.getGameData()
		if giftData == nil {
			continue
		}
		if giftData.Reward == treasureModelID {
			return gift
		}
	}
	return nil
}

func (lgm *limitGiftMgrSt) getLimitGiftByGoodsID(goodsID string) iLimitGift {
	for _, gift := range lgm.gifts {
		giftData := gift.getGameData()
		if giftData == nil {
			continue
		}
		if giftData.GiftID == goodsID {
			return gift
		}
	}
	return nil
}

func (lgm *limitGiftMgrSt) heartBeat(now time.Time) {
	for _, gift := range lgm.gifts {
		gift.tryRefresh(now)
	}
}

func (lgm *limitGiftMgrSt) onMaxPvpLevelUp(limitGiftGameData gamedata.ILimitGiftGameData, maxPvpLevel int) {
	if limitGiftGameData == nil {
		return
	}

	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	maxTeam := rankGameData.Ranks[maxPvpLevel].Team
	//var oldTeam int
	//if maxPvpLevel > 1 {
	//	oldTeam = rankGameData.Ranks[maxPvpLevel - 1].Team
	//}
	//if maxTeam <= oldTeam {
	//	return
	//}

	isUpdate := false
	allLimitGifts := limitGiftGameData.GetAllLimitGifts(lgm.player.GetArea())
L1:	for _, giftData := range allLimitGifts {
		var showTeam int
		for _, conditionInfo := range giftData.ShowConditions {
			if len(conditionInfo) >= 2 && conditionInfo[0] == "maxTeam" {
				showTeam, _ = strconv.Atoi(conditionInfo[1])
				if showTeam > maxTeam {
					continue L1
				}
			}
		}

		isExist := false
		for _, gift := range lgm.gifts {
			if gift.getGiftID() == giftData.GiftID {
				isExist = true
				break
			}
		}

		if !isExist {
			gift, attr := lgm.newLimitGift(lgm.player, giftData)
			if attr != nil {
				lgm.limitGiftsAttr.AppendMapAttr(attr)
				isUpdate = true
				if gift != nil {
					lgm.gifts = append(lgm.gifts, gift)
				}
			}
		}
	}

	if isUpdate {
		lgm.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_LimitGift)
	}
	lgm.caclHint(false)
}

func (lgm *limitGiftMgrSt) caclHint(isLogin bool) {
	var count int
	lgm.forEachCanShopShowLimitGifts(func(gift iLimitGift) {
		if gift.isNew() {
			count++
		}
	})

	if count > 0 {
		if isLogin {
			lgm.player.AddHint(pb.HintType_HtLimitGift, count)
		} else {
			lgm.player.UpdateHint(pb.HintType_HtLimitGift, count)
		}
	} else {
		lgm.player.DelHint(pb.HintType_HtLimitGift)
	}
}

func (lgm *limitGiftMgrSt) onReborn() {
	var needNotify bool
	for _, gift := range lgm.gifts {
		needNotify2 := gift.onReborn()
		if needNotify2 {
			needNotify = needNotify2
		}
	}

	if needNotify {
		lgm.caclHint(false)
	}
}

func (lgm *limitGiftMgrSt) newLimitGiftByAttr(attr *attribute.MapAttr, player types.IPlayer) (lg iLimitGift, hasNew bool) {
	giftID := attr.GetStr("giftID")
	if giftID == "1ygacha" {
		/*
		gift := &gaChaGift{}
		gift.player = player
		gift.attr = attr
		gift.hideConditions = nil
		gift.showConditions = []limitGiftCondition{}
		gift.buyConditions = []limitGiftCondition{}
		lg = gift
		*/
		return nil, false
	} else {
		gift := &limitGift{
			player: player,
			attr:   attr,
			hideConditions: nil,
			showConditions: []limitGiftCondition{},
			buyConditions: []limitGiftCondition{},
		}

		key := gift.genOldRewardKey()
		if attr.GetBool(key) {
			// 兼容老数据
			attr.Del(key)
			attr.SetInt(gift.genRewardKey(), 1)
		}
		lg = gift
	}

	giftData := lg.getGameData()
	if giftData == nil {
		return lg, true
	}

	lg.setVersion(giftData.Version)
	for _, conditionInfo := range giftData.BuyConditions {
		c := lgm.newCondition(conditionInfo)
		if c != nil {
			lg.addBuyCondition(c)
		}
	}

	for _, conditionInfo := range giftData.ShowConditions {
		c := lgm.newCondition(conditionInfo)
		if c != nil {
			lg.addShowCondition(c)
		}
	}

	for _, conditionInfo := range giftData.HideConditions {
		c := lgm.newCondition(conditionInfo)
		if c != nil {
			lg.addHideCondition(c)
		}
	}

	if giftData.NewbieGiftIDPrefix != "" {
		teamGift, ok := lgm.prefix2TeamGift[giftData.NewbieGiftIDPrefix]
		if !ok {
			teamGift = &teamGiftSt{}
			lgm.prefix2TeamGift[giftData.NewbieGiftIDPrefix] = teamGift
			teamGift.addRawGift(lg)
			return teamGift, true
		} else {
			curGift := teamGift.curGift
			teamGift.addRawGift(lg)
			if teamGift.curGift != curGift {
				hasNew = true
			}
			return nil, hasNew
		}
	}

	return lg, true
}

func (lgm *limitGiftMgrSt) newLimitGift(player types.IPlayer, giftData *gamedata.LimitGift) (iLimitGift, *attribute.MapAttr) {
	attr := attribute.NewMapAttr()
	attr.SetStr("giftID", giftData.GiftID)
	attr.SetBool("isNew", true)
	if giftData.ContinueTime > 0 {
		attr.SetInt64("tmout", time.Now().Unix() + giftData.ContinueTime)
	}
	attr.SetInt("version", giftData.Version)
	gift, hasNew := lgm.newLimitGiftByAttr(attr, player)
	if gift != nil {
		return gift, attr
	}
	if !hasNew {
		attr = nil
	}
	return nil, attr
}


type limitGift struct {
	player types.IPlayer
	attr *attribute.MapAttr

	// 为nil时，永远不满足条件
	showConditions []limitGiftCondition
	hideConditions []limitGiftCondition
	buyConditions []limitGiftCondition
}

func (lg *limitGift) onCrossDay() {}
func (lg *limitGift) onLogin() {}

func (lg *limitGift) setVersion(version int) {
	lg.attr.SetInt("version", version)
}

func (lg *limitGift) onReborn() bool {
	giftData := lg.getGameData()
	if giftData != nil && (giftData.NewbieGiftIDPrefix != "" || giftData.CommonGiftIDPrefix != "") && lg.isBuy() {
		lg.reset()
		return true
	}
	return false
}

func (lg *limitGift) reset() {
	lg.attr.Del(lg.genRewardKey())
	lg.setIsNew(true)
	lg.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_LimitGift)
}

func (lg *limitGift) getVersion() int {
	return lg.attr.GetInt("version")
}

func (lg *limitGift) String() string {
	return fmt.Sprintf("[limitGift %s, isTimeout=%s, isBuy=%s]", lg.getGiftID(), lg.isTimeout(),
		lg.isBuy())
}

func (lg *limitGift) getGiftID() string {
	return lg.attr.GetStr("giftID")
}

func (lg *limitGift) getTimeout() int64 {
	return lg.attr.GetInt64("tmout")
}

func (lg *limitGift) setTimeout(t int64) {
	lg.attr.SetInt64("tmout", t)
}

func (lg *limitGift) isTimeout() bool {
	timeout := lg.getTimeout()
	if timeout > 0 {
		return time.Now().Unix() >= timeout
	} else {
		return false
	}
}

func (lg *limitGift) isBuy() bool {
	giftData := lg.getGameData()
	if giftData == nil {
		return true
	}

	if giftData.BuyLimitCnt <= 0 {
		return false
	}

	if giftData.GiftID == minVipCardGiftID {
		st := module.OutStatus.GetStatus(lg.player, consts.OtMinVipCard)
		if st != nil {
			return true
		}
		return false
	}
	return lg.attr.GetInt(lg.genRewardKey()) >= giftData.BuyLimitCnt
}

func (lg *limitGift) canBuy() bool {
	if lg.isBuy() {
		return false
	}
	if lg.isTimeout() {
		return false
	}
	giftData := lg.getGameData()
	if giftData == nil {
		return false
	}

	return lg.checkCondition(lg.buyConditions) && lg.checkCondition(lg.showConditions) && !lg.checkCondition(lg.hideConditions)
}

func (lg *limitGift) canShow() bool {
	giftData := lg.getGameData()
	if giftData == nil {
		return false
	}

	loc, _ := time.LoadLocation("Local")
	if giftData.BeginTime != "" {
		beginTime, _ := time.ParseInLocation(timeLayout, giftData.BeginTime, loc)
		if beginTime.Unix() > time.Now().Unix() {
			return false
		}
	}
	if giftData.EndTime != "" {
		endTime, _ := time.ParseInLocation(timeLayout, giftData.EndTime, loc)
		if endTime.Unix() <= time.Now().Unix() {
			return false
		}
	}

	return lg.checkCondition(lg.showConditions) && !lg.checkCondition(lg.hideConditions)
}

func (lg *limitGift) genRewardKey() string {
	ver := lg.getVersion()
	giftData := lg.getGameData()
	var rewardTreasure string
	if giftData != nil {
		rewardTreasure = giftData.Reward
	}

	return fmt.Sprintf("n_%s_%d", rewardTreasure, ver)
}

func (lg *limitGift) genOldRewardKey() string {
	// 兼容老数据
	ver := lg.getVersion()
	giftData := lg.getGameData()
	var rewardTreasure string
	if giftData != nil {
		rewardTreasure = giftData.Reward
	}
	if ver <= 0 {
		return rewardTreasure
	} else {
		return fmt.Sprintf("%s_%d", rewardTreasure, ver)
	}
}

func (lg *limitGift) onBuy() {
	rewardKey := lg.genRewardKey()
	cnt := lg.attr.GetInt(rewardKey) + 1
	lg.attr.SetInt(rewardKey, cnt)
	giftData := lg.getGameData()
	if giftData.BuyLimitCnt > 0 && cnt >= giftData.BuyLimitCnt && giftData.RefreshTime > 0 {
		lg.setTimeout(time.Now().Unix())
	}
}

func (lg *limitGift) buy() (reply proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType) {
	giftData := lg.getGameData()
	if giftData == nil {
		return
	}
	if lg.isBuy() {
		return
	}

	lg.onBuy()
	reply2 := &pb.BuyLimitGiftReply{}
	goodsType = pb.SdkRechargeResult_Unknow

	if giftData.Reward != "" {
		reply2.GiftReward = lg.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
			giftData.Reward, false)
		goodsType = pb.SdkRechargeResult_LimitGift
	} else if giftData.RewardTbl != "" {
		rr := module.Reward.GiveReward(lg.player, giftData.RewardTbl)
		if rr != nil {
			reply2.Privileges = rr.GetPrivileges()
			goodsType = pb.SdkRechargeResult_LimitGift
		}
	}

	glog.Infof("limitGift buy uid=%d, giftID=%s, reward=%s, privileges=%v", lg.player.GetUid(), giftData.GiftID,
		giftData.Reward, reply2.Privileges)

	if giftData.GiftID == minVipCardGiftID {
		st := module.OutStatus.GetStatus(lg.player, consts.OtMinVipCard)
		if st == nil {
			module.OutStatus.AddStatus(lg.player, consts.OtMinVipCard, 7 * 24 * 3600)
		} else {
			st.Over(7 * 24 * 3600)
		}
	}

	reply = reply2

	evq.CallLater(func() {
		lg.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_LimitGift)
		lg.player.GetComponent(consts.ShopCpt).(*shopComponent).limitGiftMgr.caclHint(false)
	})

	return
}

func (lg *limitGift) getGameData() *gamedata.LimitGift {
	gdata := lg.player.GetComponent(consts.ShopCpt).(*shopComponent).getLimitGiftGameData()
	if gdata == nil {
		return nil
	}
	return gdata.GetAreaGift(lg.player.GetArea(), lg.getGiftID())
}

func (lg *limitGift) tryRefresh(now time.Time) {
	giftData := lg.getGameData()
	if giftData == nil {
		return
	}
	if giftData.RefreshTime <= 0 {
		return
	}

	if !lg.isBuy() || !lg.isTimeout() {
		return
	}

	nowTS := now.Unix()
	if nowTS < lg.getTimeout() + giftData.RefreshTime {
		return
	}

	if giftData.ContinueTime > 0 {
		lg.attr.SetInt64("tmout", nowTS + giftData.ContinueTime)
	}
	lg.reset()
}

func (lg *limitGift) packMsg() *pb.LimitGift {
	giftData := lg.getGameData()
	if giftData == nil {
		return nil
	}
	if lg.isTimeout() {
		return nil
	}

	return &pb.LimitGift{
		GiftID: lg.getGiftID(),
		//RemainTime: int32(remainTime),
		Price: int32(giftData.JadePrice),
		IsNew: lg.isNew(),
	}
}

func (lg *limitGift) isNew() bool {
	if lg.attr.GetBool("isNew") {
		data := lg.getGameData()
		if data == nil || data.BuyLimitCnt <= 0 {
			return false
		} else {
			return true
		}
	} else {
		return false
	}
}

func (lg *limitGift) setIsNew(val bool) {
	if val {
		lg.attr.SetBool("isNew", true)
	} else {
		lg.attr.Del("isNew")
	}
}

func (lg *limitGift) addShowCondition(condition limitGiftCondition) {
	lg.showConditions = append(lg.showConditions, condition)
}

func (lg *limitGift) addHideCondition(condition limitGiftCondition) {
	if lg.hideConditions == nil {
		lg.hideConditions = []limitGiftCondition{}
	}
	lg.hideConditions = append(lg.hideConditions, condition)
}

func (lg *limitGift) addBuyCondition(condition limitGiftCondition) {
	lg.buyConditions = append(lg.buyConditions, condition)
}

func (lg *limitGift) checkCondition(conditions []limitGiftCondition) bool {
	if conditions == nil {
		return false
	}
	for _, c := range conditions {
		if !c(lg.player) {
			return false
		}
	}
	return true
}


type gaChaGift struct {
	limitGift
}

func (gc *gaChaGift) canBuy() bool {
	if !gc.limitGift.canBuy() {
		return false
	}
	return gc.getBuyCnt() <= 0
}

func (gc *gaChaGift) buy() (reply proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType) {
	giftData := gc.getGameData()
	if giftData == nil {
		return
	}

	if !gc.canBuy() {
		return
	}

	reply2 := &pb.RechargeLotteryReply{}

	result := module.Reward.GiveReward(gc.player, giftData.RewardTbl)
	if result != nil && len(result.GetRewardIdxs()) > 0 {
		reply2.RewardIdx = int32(result.GetRewardIdxs()[0])
		if len(result.GetPrivileges()) > 0 {
			reply2.Privileges = result.GetPrivileges()[0]
		}
	}

	glog.Infof("oneyGaCha buy uid=%d, giftID=%s, rewardId=%d, privileges=%v", gc.player.GetUid(), giftData.GiftID,
		reply2.RewardIdx, reply2.Privileges)

	timeout := gc.getNextTime()

	gc.setTimeout(timeout)
	gc.setBuyCnt(1)
	gc.caclHint(false)
	return reply2, pb.SdkRechargeResult_Lottery
}

func (gc *gaChaGift) setBuyCnt(value int) {
	gc.attr.SetInt("buyCnt", value)
}

func (gc *gaChaGift) getBuyCnt() int{
	return gc.attr.GetInt("buyCnt")
}

func (gc *gaChaGift) setTimeout(value int32) {
	gc.attr.SetInt32("timeout", value)
}

func (gc *gaChaGift) getTimeout() int32 {
	return gc.attr.GetInt32("timeout")
}

func (gc *gaChaGift) getNextTime() int32 {
	now := time.Now()
	nextTime := time.Date(now.Year(), now.Month(), now.Day(), 24, 0, 0, 0, now.Location())
	if !now.Before(nextTime) {
		nextTime = nextTime.Add(86400 * time.Second)
	}
	return int32(nextTime.Unix())
}

func (gc *gaChaGift) caclHint(isLogin bool){
	if !gc.canBuy(){
		gc.player.DelHint(pb.HintType_HtGaCha)
	}else {
		if isLogin {
			gc.player.AddHint(pb.HintType_HtGaCha, 1)
		}else{
			gc.player.UpdateHint(pb.HintType_HtGaCha, 1)
		}
	}
}

func (gc *gaChaGift) onLogin() {
	curTime :=time.Now().Unix()
	if int32(curTime) > gc.getTimeout() {
		gc.setBuyCnt(0)
	}
	gc.caclHint(true)
}

func (gc *gaChaGift) onCrossDay(){
	timeout := gc.getNextTime()
	gc.setTimeout(timeout)
	gc.setBuyCnt(0)
	gc.caclHint(false)
}

func (gc *gaChaGift) setIsNew(val bool) {

}

// 某个国家的段位礼包，gift + camp 开头的
type teamGiftSt struct {
	curGift iLimitGift  // 当前应该显示的
	allGifts []iLimitGift
}

func (lg *teamGiftSt) buy() (reply proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType) {
	if lg.curGift == nil {
		return
	}

	reply, goodsType = lg.curGift.buy()
	lg.chooseShowGift()
	return
}

func (lg *teamGiftSt) getGameData() *gamedata.LimitGift {
	if lg.curGift == nil {
		return nil
	}
	return lg.curGift.getGameData()
}

func (lg *teamGiftSt) tryRefresh(now time.Time) {
	if lg.curGift == nil {
		return
	}
	lg.curGift.tryRefresh(now)
}

func (lg *teamGiftSt) packMsg() *pb.LimitGift {
	if lg.curGift == nil {
		return nil
	}
	return lg.curGift.packMsg()
}

func (lg *teamGiftSt) getGiftID() string {
	if lg.curGift == nil {
		return ""
	}
	return lg.curGift.getGiftID()
}

func (lg *teamGiftSt) setIsNew(val bool) {
	if lg.curGift == nil {
		return
	}
	lg.curGift.setIsNew(val)
}

func (lg *teamGiftSt) isBuy() bool {
	if lg.curGift == nil {
		return true
	}
	return lg.curGift.isBuy()
}

func (lg *teamGiftSt) canShow() bool {
	if lg.curGift == nil {
		return false
	}
	return lg.curGift.canShow()
}

func (lg *teamGiftSt) isNew() bool {
	if lg.curGift == nil {
		return false
	}
	return lg.curGift.isNew()
}

func (lg *teamGiftSt) onReborn() bool {
	var ok bool
	for _, g := range lg.allGifts {
		if g.onReborn() {
			ok = true
		}
	}

	if ok {
		lg.chooseShowGift()
	}

	return ok
}

func (lg *teamGiftSt) reset() {
	for _, g := range lg.allGifts {
		g.reset()
	}
}

func (lg *teamGiftSt) canBuy() bool {
	if lg.curGift == nil {
		return false
	}
	return lg.curGift.canBuy()
}

func (lg *teamGiftSt) onCrossDay() {
	for _, g := range lg.allGifts {
		g.onCrossDay()
	}
}

func (lg *teamGiftSt) onLogin() {
	for _, g := range lg.allGifts {
		g.onLogin()
	}
}

func (lg *teamGiftSt) setVersion(version int) {
	for _, g := range lg.allGifts {
		g.setVersion(version)
	}
}

func (lg *teamGiftSt) addShowCondition(condition limitGiftCondition) {
	for _, g := range lg.allGifts {
		g.addShowCondition(condition)
	}
}

func (lg *teamGiftSt) addHideCondition(condition limitGiftCondition) {
	for _, g := range lg.allGifts {
		g.addHideCondition(condition)
	}
}

func (lg *teamGiftSt) addBuyCondition(condition limitGiftCondition) {
	for _, g := range lg.allGifts {
		g.addBuyCondition(condition)
	}
}

func (lg *teamGiftSt) addRawGift(g iLimitGift) {
	for _, g2 := range lg.allGifts {
		if g2 == g || g2.getGiftID() == g.getGiftID() {
			return
		}
	}
	lg.allGifts = append(lg.allGifts, g)
	lg.chooseShowGift()
}

func (lg *teamGiftSt) chooseShowGift() {
	var curData *gamedata.LimitGift
	//oldGift := lg.curGift
	var curCanBuy bool
	if lg.curGift != nil {
		curData = lg.curGift.getGameData()
		curCanBuy = lg.curGift.canBuy() && lg.curGift.canShow()
	}

	for _, g := range lg.allGifts {
		if lg.curGift == nil {
			lg.curGift = g
			curData = g.getGameData()
			curCanBuy = g.canBuy() && g.canShow()
			continue
		}

		if !g.canBuy() || !g.canShow() {
			if !curCanBuy {
				data := g.getGameData()
				if data.MaxTeam > curData.MaxTeam {
					lg.curGift = g
					curData = data
				}
			}
			continue
		}

		if curData == nil {
			lg.curGift = g
			curData = g.getGameData()
		} else {
			data := g.getGameData()
			if data.MaxTeam > curData.MaxTeam || !curCanBuy {
				lg.curGift = g
				curData = data
			}
		}
		curCanBuy = true
	}

	//if oldGift != lg.curGift && lg.curGift != nil {
	//	lg.curGift.setIsNew(true)
	//}
}
