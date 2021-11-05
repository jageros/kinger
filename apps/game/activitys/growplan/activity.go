package growplan

import (
	"kinger/gopuppy/common/glog"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"strings"
)

const (
	ty_unknowType_           = 0  //未知类型
	ty_useCampCardWinBattle_ = 1  //使用{0}级卡组对战胜利{1}场（{2}/{1}）
	ty_achieveLevel_         = 2  //达到#cy{0}#n
	ty_hasLevelCard_         = 3  //拥有{0}个{1}级卡牌（{2}/{0}）
	ty_hasStarCard_          = 4  //拥有{0}个{1}星卡牌（{2}/{0}）
	ty_hasFriends_           = 5  //拥有{0}名好友（{1}/{0}）
	ty_openTreasure_         = 6  //开启{0}个宝箱（{1}/{0}）
	ty_jadeConsume_          = 7  //消耗{0}宝玉（{1}/{0}）
	ty_combat_               = 8  //约战{0}次（{1}/{0}）
	ty_watchBattleReport_    = 9  //观看{0}场战报（{1}/{0}）
	ty_sendBattleReport_     = 10 //发布{0}场战报（{1}/{0}）
	ty_hitOutStarCard_       = 11 //打出{0}次{1}星卡牌（{2}/{0}）
	ty_hitOutCampCard_       = 12 //打出{0}次{1}卡牌（{2}/{0}）
	ty_useCampWin_           = 13 //使用{0}天梯对战胜利{1}次（{2}/{1}）
	ty_useCampBattle_        = 14 //使用{0}天梯对战{1}次（{2}/{1}）
	ty_finshMission_         = 15 //{0}个任务（{1}/{0}）
	ty_totalRecharge_        = 16 //累计充值{0}元({1}/{2}
	ty_totalBuyJade_         = 17 //累计购买{0}宝玉({1}/{2})
	ty_vipExclusive_         = 18 //月卡专属奖励
	ty_login_                = 19 //{0}期间登陆游戏
	ty_battleNum_            = 20 //天梯对战{0}次
	ty_battlewinNum          = 21 //天梯对战胜利{0}次
	ty_BuyLimtGift_          = 22 //购买后领取
	ty_ContinuousWinNum_     = 23 //在{0}连续胜利{1}场

)

var mod *activity

type activity struct {
	aTypes.BaseActivity
	id2Reward map[int]*gamedata.ActivityGrowPlanRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfGrowPlan)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("GrowPlan activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityGrowPlanRewardGameData)
	}
}

func (a *activity) getRewardIdList(aid int) []int {
	var idList []int
	if rw, ok := a.id2Reward[aid]; ok {
		for k, _ := range rw.ID2ActivityGrowPlanReward {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(aid, rid int) (int, []int, error) {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityGrowPlanReward[rid]; ok {
			return rw.ConditionType, rw.ConditionVal, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	glog.Errorf("get reward condition err=%s, activityID=%d, rewardID=%d", err)
	return 0, nil, err
}

func (a *activity) getGoodsId(aid, rid int) string {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityGrowPlanReward[rid]; ok {
			return rw.Purchase
		}
	}
	return ""
}

func (a *activity) getRewardData(aid, rid int) *gamedata.ActivityGrowPlanReward {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityGrowPlanReward[rid]; ok {
			return rw
		}
	}
	return nil
}

func (a *activity) getRewardMap(aid, rid int) map[string]int32 {
	rw := map[string]int32{}
	rd := a.getRewardData(aid, rid)
	if rd == nil {
		err := gamedata.GameError(aTypes.GetRewardError)
		glog.Errorf("GrowPlan activity getRewardMap get reward data err=%s, activityID=%d, rewardID=%d", err, aid, rid)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("GrowPlan activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, aid, rid, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("GrowPlan activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, aid, rid, strl[1])
			return nil
		}

		rw[strl[0]] = int32(num)
	}
	return rw
}

func onFightEnd(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	isWin := args[1].(bool)
	battleData := args[2].(*pb.EndFighterData)

	camp := int(battleData.Camp)
	skCards := battleData.InitHandCards
	hitOUtCard := battleData.UseCards
	ids := mod.IAMod.GetActivityIdList()
	isSameLevelCard := true
	cardPool := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	lvl := cardPool.GetCardByGid(skCards[0].GCardID).Level
	if isWin {
		for _, skCard := range skCards {
			clvl := cardPool.GetCardByGid(skCard.GCardID).Level
			if lvl != clvl {
				isSameLevelCard = false
				break
			}
		}
	}

	for _, id := range ids {
		if !p.ipc.Conform(id) {
			continue
		}
		p.setContinuousWinNum(id, isWin)
		if isWin {
			num := p.get2ArgAttrNum(id, aTypes.GrowPlan_useCampWinNum, camp) + 1
			p.set2ArgAttrNum(id, aTypes.GrowPlan_useCampWinNum, camp, num)
			num0 := p.get2ArgAttrNum(id, aTypes.GrowPlan_useCampWinNum, 0) + 1
			p.set2ArgAttrNum(id, aTypes.GrowPlan_useCampWinNum, 0, num0)

			if isSameLevelCard {
				num := p.get2ArgAttrNum(id, aTypes.GrowPlan_campCardWin, lvl) + 1
				p.set2ArgAttrNum(id, aTypes.GrowPlan_campCardWin, lvl, num)
				num0 := p.get2ArgAttrNum(id, aTypes.GrowPlan_campCardWin, 0) + 1
				p.set2ArgAttrNum(id, aTypes.GrowPlan_campCardWin, 0, num0)
			}

		}
		num := p.get2ArgAttrNum(id, aTypes.GrowPlan_useCampBattleNum, camp) + 1
		p.set2ArgAttrNum(id, aTypes.GrowPlan_useCampBattleNum, camp, num)
		num0 := p.get2ArgAttrNum(id, aTypes.GrowPlan_useCampBattleNum, 0) + 1
		p.set2ArgAttrNum(id, aTypes.GrowPlan_useCampBattleNum, 0, num0)

		for _, gid := range hitOUtCard {
			cmp := cardPool.GetCardByGid(gid).Camp
			num1 := p.get2ArgAttrNum(id, aTypes.GrowPlan_hitOutCampCardNum, cmp) + 1
			p.set2ArgAttrNum(id, aTypes.GrowPlan_hitOutCampCardNum, cmp, num1)
			num10 := p.get2ArgAttrNum(id, aTypes.GrowPlan_hitOutCampCardNum, 0) + 1
			p.set2ArgAttrNum(id, aTypes.GrowPlan_hitOutCampCardNum, 0, num10)

			star := cardPool.GetCardByGid(gid).Rare
			num2 := p.get2ArgAttrNum(id, aTypes.GrowPlan_hitOutStarCardNum, star) + 1
			p.set2ArgAttrNum(id, aTypes.GrowPlan_hitOutStarCardNum, star, num2)
			num20 := p.get2ArgAttrNum(id, aTypes.GrowPlan_hitOutStarCardNum, 0) + 1
			p.set2ArgAttrNum(id, aTypes.GrowPlan_hitOutStarCardNum, 0, num20)
		}
		p.pushFinshNum(id, ty_unknowType_)
	}
	p.updateHint()
}

func onConsume(args ...interface{}) {
	player := args[0].(types.IPlayer)
	resType := args[1].(int)
	if resType != consts.Jade {
		return
	}
	money := args[3].(int)
	if money >= 0 {
		return
	}
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			num := p.get1ArgAttrNum(id, aTypes.GrowPlan_jadeConsume) + (-money)
			p.set1ArgAttrNum(id, aTypes.GrowPlan_jadeConsume, num)
			p.pushFinshNum(id, ty_jadeConsume_)
		}
	}
	p.updateHint()
}

func onOpenTreasure(args ...interface{}) {
	player := args[0].(types.IPlayer)
	isDaily := args[1].(bool)
	if isDaily {
		return
	}
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			num := p.get1ArgAttrNum(id, aTypes.GrowPlan_TreasureOpenNum) + 1
			p.set1ArgAttrNum(id, aTypes.GrowPlan_TreasureOpenNum, num)
			p.pushFinshNum(id, ty_openTreasure_)
		}
	}
	p.updateHint()
}
func onCombat(args ... interface{}) {
	for _, arg := range args {
		player, ok := arg.(types.IPlayer)
		if !ok {
			continue
		}
		p := newComponent(player)
		for _, id := range mod.IAMod.GetActivityIdList() {
			if p.ipc.Conform(id) {
				num := p.get1ArgAttrNum(id, aTypes.GrowPlan_combat) + 1
				p.set1ArgAttrNum(id, aTypes.GrowPlan_combat, num)
				p.pushFinshNum(id, ty_combat_)
			}
		}
		p.updateHint()
	}
}

func onFinshMission(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			num := p.get1ArgAttrNum(id, aTypes.GrowPlan_finshMissionNum) + 1
			p.set1ArgAttrNum(id, aTypes.GrowPlan_finshMissionNum, num)
			p.pushFinshNum(id, ty_finshMission_)
		}
	}
	p.updateHint()
}

func onShareBattleReport(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			num := p.get1ArgAttrNum(id, aTypes.GrowPlan_sendBattleReportNum) + 1
			p.set1ArgAttrNum(id, aTypes.GrowPlan_sendBattleReportNum, num)
			p.pushFinshNum(id, ty_sendBattleReport_)
		}
	}
	p.updateHint()
}

func onWatchBattleReport(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			num := p.get1ArgAttrNum(id, aTypes.GrowPlan_watchBattleReportNum) + 1
			p.set1ArgAttrNum(id, aTypes.GrowPlan_watchBattleReportNum, num)
			p.pushFinshNum(id, ty_watchBattleReport_)
		}
	}
	p.updateHint()
}

func onMaxPvpLevelUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			p.pushFinshNum(id, ty_achieveLevel_)
		}
	}
	p.updateHint()
}

func onAddFriend(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			p.pushFinshNum(id, ty_hasFriends_)
		}
	}
	p.updateHint()
}

func onCardUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			p.pushFinshNum(id, ty_hasLevelCard_)
			p.pushFinshNum(id, ty_hasStarCard_)
		}
	}
	p.updateHint()
}

func onRecharge(args ...interface{}) {
	player := args[0].(types.IPlayer)
	money := args[1].(int)
	goodsId := args[2].(string)
	p := newComponent(player)
	for _, aid := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(aid) {
			mNum := p.get1ArgAttrNum(aid, aTypes.GrowPlan_totalRecharge)
			p.set1ArgAttrNum(aid, aTypes.GrowPlan_totalRecharge, mNum + money)
			for _, rid := range mod.getRewardIdList(aid) {
				if mod.getGoodsId(aid, rid) == goodsId {
					p.setBuyGift(aid, goodsId)
					module.Televise.SendNotice(pb.TeleviseEnum_BuyGrowPlan, player.GetName())
					break
				}
			}
		}
	}
	p.updateHint()
}

func updateAllPlayerHint() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := newComponent(player)
		p.updateHint()
	})
}
